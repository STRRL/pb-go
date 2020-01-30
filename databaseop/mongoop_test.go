package databaseop

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"testing"
	"time"
)

var globalMGC *mongo.Client

func TestMongoDBConn(t *testing.T) {
	var mgcli = MongoDB{
		DbConn:         globalMGC,
		DbURI:          "mongodb://localhost:27017",
		DbColl:         mongo.Collection{},
		DefaultDB:      "pbgo",
		DefaultColl:    "userdata",
		DefaultTimeout: time.Time{},
		BsonRData:      make(chan UserData, 100),
		BsonWData:      make(chan UserData, 100),
	}
	clientOptions := options.Client()
	clientOptions.ApplyURI(mgcli.DbURI)
	clientOptions.SetMinPoolSize(2)
	clientOptions.SetMaxPoolSize(4)
	err := mgcli.connNCheck(clientOptions)
	mgcli.DbColl = *mgcli.DbConn.Database(mgcli.DefaultDB).Collection(mgcli.DefaultColl)
	if err != nil {
		t.Fail()
	}
	var tempIP string
	tempIP, err = IP2Intstr("113.55.13.1")
	if err != nil {
		t.Fail()
	}
	var IPval primitive.Decimal128
	IPval, _ = primitive.ParseDecimal128(tempIP)
	var UserDT primitive.DateTime
	UserDT = primitive.NewDateTimeFromTime(time.Now().Add(24 * time.Hour))
	testdt1 := UserData{
		WaitVerify: true,
		ShortId:    "2s4D",
		UserIP:     IPval,
		ExpireAt:   UserDT,
		Data:       Pack2BinData("testdata001"),
		PwdIsSet:   true,
		Password:   "He1loWorld234",
	}
	mgcli.BsonWData <- testdt1
	err = mgcli.itemCreate()
	if err != nil {
		log.Println("Failed to create document")
		t.Fail()
	}
	filter1 := bson.M{"shortId": "2s4D"}
	err = mgcli.itemRead(filter1)
	if err != nil {
		t.Fail()
	} else {
		queryres := <-mgcli.BsonRData
		log.Println(queryres)
	}
	time.Sleep(5 * time.Second)
	update1 := bson.D{
		{"$set", bson.D{
			{"data", Pack2BinData("testdata002")},
		}},
	}
	err = mgcli.itemUpdate(filter1, update1)
	if err != nil {
		t.Fail()
	}
	time.Sleep(5 * time.Second)
	err = mgcli.itemDelete(filter1)
	if err != nil {
		t.Fail()
	}
	log.Println("Test Done!")
	os.Exit(0)
}