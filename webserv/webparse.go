package webserv

import (
	"bytes"
	"encoding/base64"
	"github.com/pb-go/pb-go/config"
	"github.com/pb-go/pb-go/contenttools"
	"github.com/pb-go/pb-go/databaseop"
	_ "github.com/pb-go/pb-go/statik" // indirect
	"github.com/pb-go/pb-go/templates"
	"github.com/pb-go/pb-go/utils"
	"github.com/rakyll/statik/fs"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

// StatikFS Global Variable
var (
	STFS http.FileSystem
)

// InitStatikFS : Extract assets from compiled go file
func InitStatikFS(stfs *http.FileSystem) {
	var err error
	*stfs, err = fs.New()
	if err != nil {
		log.Fatalln(err)
	}
}

// UserUploadParse : Upload API Processing Function
func UserUploadParse(c *fasthttp.RequestCtx) {
	var err error
	// init obj
	var userForm = databaseop.UserData{}
	// evaluate remote ip
	var rmtIPhd string
	rmtIPhd, err = utils.IP2Intstr(string(c.Request.Header.Peek("X-Real-IP")))
	if len(rmtIPhd) < 1 || err != nil {
		c.SetStatusCode(http.StatusBadGateway)
		return
	}
	// first parse user form
	var userExpire int
	userPwd := c.FormValue("p") // password can be nil and will get a hash, so no worry.
	userExpireB := c.FormValue("e")
	if userExpireB == nil {
		userExpire = config.ServConf.Content.ExpireHrs
	} else {
		userExpire, err = strconv.Atoi(string(userExpireB))
		if err != nil {
			c.SetStatusCode(http.StatusBadRequest)
			return
		}
	}
	var userData []byte
	userPOSTFdHd, err := c.FormFile("d")
	if err != nil || userPOSTFdHd == nil {
		userData = c.FormValue("d")
		if userData == nil {
			c.SetStatusCode(http.StatusBadRequest)
			return
		}
	} else {
		userPOSTFile, err := userPOSTFdHd.Open()
		if err != nil {
			c.SetStatusCode(http.StatusBadRequest)
			return
		}
		userDatabuf := bytes.NewBuffer(nil)
		if _, err := io.Copy(userDatabuf, userPOSTFile); err != nil {
			c.SetStatusCode(http.StatusBadRequest)
			return
		}
		userData = userDatabuf.Bytes()
		_ = userPOSTFile.Close()
	}
	// uploaded file length detect
	if len(userData) > 2*1024*1024 {
		c.SetStatusCode(http.StatusBadRequest)
		return
	}
	// set expire check
	if userExpire > config.ServConf.Content.ExpireHrs || userExpire < 0 || len(userData) < 1 {
		c.SetStatusCode(http.StatusBadRequest)
		return
	}
	// given shortid
	userForm.ShortID, _ = utils.GetNanoID()
	userForm.PwdIsSet = len(string(userPwd)) >= 1
	userForm.UserIP, _ = primitive.ParseDecimal128(rmtIPhd)
	// then detect if enable abuse detection
	if config.ServConf.Content.DetectAbuse {
		if !utils.ContentValidityCheck(userData) {
			c.SetStatusCode(http.StatusForbidden)
			return
		}
	}
	// then encrypt
	var userDataEnc []byte
	userDataEnc, userForm.Password, err = utils.EncryptData(userData, userPwd)
	if err != nil {
		c.SetStatusCode(http.StatusBadRequest)
		return
	}
	userForm.Data = utils.Pack2BinData(userDataEnc)
	// calculate expire
	// if recaptcha enabled, set to 5min expires first,
	// else, set to 24hrs, then build next.
	if userExpire == 0 {
		userForm.ReadThenBurn = true
		userForm.ExpireAt = primitive.NewDateTimeFromTime(time.Now().Add(time.Duration(config.ServConf.Content.ExpireHrs) * time.Hour))
	} else {
		userForm.ExpireAt = primitive.NewDateTimeFromTime(time.Now().Add(time.Duration(userExpire) * time.Hour))
	}
	if config.ServConf.Recaptcha.Enable {
		userForm.WaitVerify = true
		// then return recaptcha url, set id param in url using rawurl_b64.
		tempurlid := base64.RawURLEncoding.EncodeToString([]byte(userForm.ShortID))
		err = databaseop.GlobalMDBC.ItemCreate(userForm)
		if err != nil {
			c.SetStatusCode(http.StatusBadGateway)
			return
		}
		redirect2URI := "/showVerify?id=" + tempurlid
		c.Redirect(redirect2URI, http.StatusFound) // use 302, instead of 307.
		c.SetBodyString("Please go to \n\n https://" + config.ServConf.Network.Host + redirect2URI + " \n\n to finish CAPTCHA.")
		return
	}
	err = databaseop.GlobalMDBC.ItemCreate(userForm)
	if err != nil {
		c.SetStatusCode(http.StatusBadGateway)
		return
	}
	// return publish url instead
	c.SetStatusCode(http.StatusOK)
	c.SetContentType("text/plain")
	respBodyStr := "Published at \n\n https://" + config.ServConf.Network.Host + "/" + userForm.ShortID + " \n\n"
	respBodyStr += "If you have set password, please append `p=<PASSWORD>` as URI Param. \n"
	respBodyStr += "If you need raw snippet, please append `f=raw` as URI Param. \n"
	c.SetBodyString(respBodyStr)
	return
}

func setShowSnipRenderData(userdt *databaseop.UserData, ctx *fasthttp.RequestCtx, israw bool) {
	decres, err := utils.DecryptData(userdt.Data.Data, ctx.FormValue("p"))
	if err != nil {
		ctx.SetStatusCode(http.StatusForbidden)
		return
	}
	ctx.SetStatusCode(http.StatusOK)
	if israw {
		ctx.SetContentType("text/plain")
		ctx.SetBody(decres)
		return
	} else {
		ctx.SetContentType("text/html; charset=utf-8")
		ctx.SetBodyString(templates.ShowSnipPageRend(string(decres)))
		return
	}
}

func readFromEmbed(statikfs http.FileSystem, filenm string, c *fasthttp.RequestCtx) {
	tempfd, err := fs.ReadFile(statikfs, filenm)
	if err != nil {
		c.SetStatusCode(http.StatusNotFound)
		return
	}
	c.SetStatusCode(http.StatusOK)
	c.SetBody(tempfd)
	return
}

// ShowSnip : Root Handler Function
func ShowSnip(c *fasthttp.RequestCtx) {
	tmpvar := c.UserValue("shortId")
	switch tmpvar {
	case nil:
		fallthrough
	case "index.html":
		c.SetContentType("text/html; charset=utf-8")
		readFromEmbed(STFS, "/index.html", c)
		return
	case "submit.html":
		c.SetContentType("text/html; charset=utf-8")
		c.SetStatusCode(http.StatusOK)
		c.SetBodyString(templates.ShowSubmitPage())
		return
	case "favicon.ico":
		c.SetContentType("image/vnd.microsoft.icon")
		readFromEmbed(STFS, "/favicon.ico", c)
		return
	case "showVerify":
		c.SetContentType("text/html; charset=utf-8")
		c.SetStatusCode(http.StatusOK)
		c.SetBodyString(templates.VerifyPageRend())
		return
	case "status":
		c.SetContentType("application/json")
		c.SetStatusCode(http.StatusOK)
		c.SetBody(retStatusJSON())
		return
	default:
		filter1 := bson.M{"shortId": tmpvar}
		readoutDta, err := databaseop.GlobalMDBC.ItemRead(filter1)
		if err != nil || readoutDta.WaitVerify {
			log.Println(err)
			c.SetStatusCode(http.StatusNotFound)
			return
		} else {
			var rawRender = string(c.FormValue("f")) == "raw"
			if readoutDta.PwdIsSet {
				uploadedpwd := c.FormValue("p")
				hashedupdpwd := utils.GenBlake2B(uploadedpwd)
				if hashedupdpwd != readoutDta.Password {
					c.SetStatusCode(http.StatusForbidden)
					return
				}
			}
			setShowSnipRenderData(&readoutDta, c, rawRender)
			if readoutDta.ReadThenBurn {
				_ = databaseop.GlobalMDBC.ItemDelete(filter1)
			}
			return
		}
	}
}

// DeleteSnip : Remove Snippet
func DeleteSnip(c *fasthttp.RequestCtx) {
	masterkey := string(c.Request.Header.Peek("X-Master-Key"))
	if masterkey == "" {
		c.SetStatusCode(http.StatusForbidden)
		return
	}
	legalkey := utils.GetUTCTimeHash(config.ServConf.Security.MasterKey)
	if masterkey == legalkey {
		curshortid := string(c.FormValue("id"))
		filter1 := bson.M{"shortId": curshortid}
		err := databaseop.GlobalMDBC.ItemDelete(filter1)
		if err != nil {
			c.SetStatusCode(http.StatusBadRequest)
			return
		} else {
			c.SetStatusCode(http.StatusAccepted)
			return
		}
	} else {
		c.SetStatusCode(http.StatusForbidden)
		return
	}
}

// StartVerifyCAPT : Accept Captcha Token and do SSV
func StartVerifyCAPT(c *fasthttp.RequestCtx) {
	if !config.ServConf.Recaptcha.Enable {
		c.SetStatusCode(http.StatusForbidden)
		return
	}
	var formsnipid []byte
	// this buffer length must be equal with the generated nanoid length,
	// otherwise the mongodb-go-driver will not find your document
	oriSnipid := c.FormValue("snipid")
	formsnipid = make([]byte, base64.RawStdEncoding.DecodedLen(len(oriSnipid)))
	_, err := base64.RawURLEncoding.Decode(formsnipid, oriSnipid)
	currentSnipid := string(formsnipid)
	if err != nil || currentSnipid == "" || len(currentSnipid) != 4 {
		c.SetStatusCode(http.StatusBadRequest)
		return
	}
	rmtIPhd := string(c.Request.Header.Peek("X-Real-IP"))
	if rmtIPhd == "" {
		c.SetStatusCode(http.StatusBadGateway)
		return
	}
	res, err := contenttools.VerifyRecaptchaResp(string(c.FormValue("g-recaptcha-response")), rmtIPhd)
	if err != nil || res == false {
		c.SetStatusCode(http.StatusForbidden)
		return
	}
	if res == true {
		filter1 := bson.M{"shortId": currentSnipid}
		update1 := bson.D{
			{"$set", bson.D{
				{"waitVerify", false},
			}},
		}
		err = databaseop.GlobalMDBC.ItemUpdate(filter1, update1)
		if err != nil {
			log.Println(err)
			c.SetStatusCode(http.StatusGone)
			return
		} else {
			c.SetStatusCode(http.StatusOK)
			c.SetContentType("text/plain")
			respBodyStr := "Verification Passed. Go to https://" + config.ServConf.Network.Host + "/" + currentSnipid + " to see your paste. \n"
			respBodyStr += "If you have set password, please append `p=<PASSWORD>` as URI Param. \n"
			respBodyStr += "If you need raw snippet, please append `f=raw` as URI Param. \n"
			c.SetBodyString(respBodyStr)
			return
		}
	}
}
