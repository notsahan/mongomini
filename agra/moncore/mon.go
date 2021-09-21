package moncore

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DefaultContextTimout time.Duration = 10 * time.Second
)

type Moncore struct {
	client *mongo.Client
}

func DefaultContext() (*context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultContextTimout)

	return &ctx, cancel
}

// MonCore is a wrapper around mongo.Client.
// URL is the connection string like 'Mongo_proto + Mongo_user + ":" + Mongo_pass + "@" + Mongo_host'
func InitMongo(url string) (*Moncore, error) {

	clientOptions := options.Client().ApplyURI(url)

	ctx_conn, cnc_conn := DefaultContext()
	defer cnc_conn()
	client, err := mongo.Connect(*ctx_conn, clientOptions)
	if err != nil {
		return nil, err
	}

	ctx_ping, cnc_ping := DefaultContext()
	defer cnc_ping()
	ping_err := client.Ping(*ctx_ping, nil)
	if ping_err != nil {
		return nil, ping_err
	}

	Print("MongoDB client connected")
	return &Moncore{client: client}, nil
}

func (MC *Moncore) Disconnect() error {

	ctx_dbr, cnc_dbr := DefaultContext()
	defer cnc_dbr()

	return MC.client.Disconnect(*ctx_dbr)
}

func (MC *Moncore) Database(name string) *Database {
	db := MC.client.Database(name)
	return &Database{db: db}
}

type Database struct {
	db *mongo.Database
}

func (MD *Database) Collection(name string) *Collection {
	return &Collection{MC: MD.db.Collection(name)}
}

func (MD *Database) ListCollectionNames() []string {

	ctx_dbr, cnc_dbr := DefaultContext()
	defer cnc_dbr()

	names, err := MD.db.ListCollectionNames(*ctx_dbr, bson.M{})

	if CheckError(err) {
		return nil
	}
	return names
}

func (MD *Database) ListCollections() map[string]*Collection {
	colnames := MD.ListCollectionNames()

	if len(colnames) == 0 {
		return map[string]*Collection{}
	}

	cols := make(map[string]*Collection, len(colnames))

	for _, colname := range colnames {
		cols[colname] = MD.Collection(colname)
	}

	return cols
}

type Collection struct {
	MC *mongo.Collection
}

// TODO : filters
func (C *Collection) Query(filter interface{}) map[string]GenericDocument {

	ctx_dbr, cnc_dbr := DefaultContext()
	defer cnc_dbr()

	qcur, qerr := C.MC.Find(*ctx_dbr, filter)

	if CheckError(qerr) {
		return nil
	}

	out := map[string]GenericDocument{}

	ctx_dbr, cnc_dbr = DefaultContext()
	defer cnc_dbr()

	for qcur.Next(*ctx_dbr) {

		// d := DBDocument_new(OutputDocTemplate)
		d := GenericDocument{}
		derr := qcur.Decode(&d)

		if !CheckError(derr) {
			out[d.ID] = d
		}

		ctx_dbr, cnc_dbr = DefaultContext()
		defer cnc_dbr()
	}

	return out

}

// Returns Inserted ID or "" if updated already existing document
func (C *Collection) SetDocument(Doc *DBDocument) string {
	ctx_dbr, cnc_dbr := DefaultContext()
	defer cnc_dbr()

	truebool := true
	res, rerr := C.MC.UpdateByID(*ctx_dbr, Doc.ID, bson.M{"$set": Doc}, &options.UpdateOptions{Upsert: &truebool})

	if CheckError(rerr) {
		return ""
	}

	return castInterfaceToString(res.UpsertedID)
}

// Returns Inserted ID or nil if updated already existing document
func (C *Collection) Set(key string, val interface{}) string {
	return C.SetDocument(&DBDocument{ID: key, Doc: val})
}

func castInterfaceToString(i interface{}) string {
	jb, je := json.Marshal(i)

	if CheckError(je) {
		return ""
	}
	oid := string(jb)

	oid = strings.TrimPrefix(oid, `"`)
	oid = strings.TrimSuffix(oid, `"`)

	if oid == "null" {
		return ""
	}

	return oid
}

// Document structure to be stored in MongoDB
type DBDocument struct {
	ID  string      `bson:"_id"`
	Doc interface{} `bson:"Doc"`
}

// Generic Document structure to be decoded into any type. Doc is a map[string]interface{}
type GenericDocument struct {
	ID  string                 `bson:"_id"`
	Doc map[string]interface{} `bson:"Doc"`
}

type Filter_MatchAll bson.M

func DBDocument_new(Doc interface{}) DBDocument {
	return DBDocument{Doc: Doc}
}
