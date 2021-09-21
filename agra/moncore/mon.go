package moncore

import (
	"context"
	"encoding/json"
	"net/http"
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

func LongRunningContext() (*context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

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
func (C *Collection) query_curser(filter *Filter) *mongo.Cursor {

	ctx_dbr, cnc_dbr := DefaultContext()
	defer cnc_dbr()

	qcur, qerr := C.MC.Find(*ctx_dbr, filter.MongoQuery)

	if CheckError(qerr) {
		return nil
	}

	return qcur
}

func (C *Collection) Query(filter *Filter) []GenericDBDocument {

	qcur := C.query_curser(filter)

	if qcur == nil {
		return nil
	}

	out := []GenericDBDocument{}

	ctx_dbr, cnc_dbr := DefaultContext()
	defer cnc_dbr()

	cerr := qcur.All(*ctx_dbr, &out)

	if CheckError(cerr) {
		return nil
	}

	return out

}

// bufferSize: 0 means unbuffered. This will load all documents at once.
func (C *Collection) QueryToChannel(filter *Filter, bufferSize int) (chan *GenericDBDocument, context.CancelFunc) {

	qcur := C.query_curser(filter)

	if qcur == nil {
		return nil, nil
	}

	out := make(chan *GenericDBDocument, bufferSize)
	ctx_dbr, cnc_dbr := LongRunningContext()

	go func() {

		for qcur.Next(*ctx_dbr) {

			d := GenericDBDocument{}
			derr := qcur.Decode(&d)

			if !CheckError(derr) {
				out <- &d
			}

		}

		close(out)

		defer cnc_dbr()
	}()

	return out, cnc_dbr

}

// Returns Inserted ID or "" if updated already existing document
func (C *Collection) SetDocument(Doc *DBDocument) WriteOperationResponse {
	ctx_dbr, cnc_dbr := DefaultContext()
	defer cnc_dbr()

	truebool := true
	res, rerr := C.MC.UpdateByID(*ctx_dbr, Doc.ID, bson.M{"$set": Doc}, &options.UpdateOptions{Upsert: &truebool})

	if CheckError(rerr) {
		return WriteOperationResponse{
			Status: 2,
			Action: "dbreq",
			Result: rerr.Error(),
		}
	}

	if res.UpsertedID == nil {
		return WriteOperationResponse{
			Status: 1,
			Action: "update",
			Result: Doc.ID,
		}
	}

	str, strOK := res.UpsertedID.(string)
	if !strOK {
		return WriteOperationResponse{
			Status: http.StatusInternalServerError,
			Action: "typecast",
			Result: "UpsertedID is not a string",
		}
	}

	return WriteOperationResponse{
		Status: 1,
		Action: "insert",
		Result: str,
	}
}

// Returns Inserted ID or nil if updated already existing document
func (C *Collection) Set(key string, val interface{}) WriteOperationResponse {
	return C.SetDocument(&DBDocument{ID: key, Doc: val})
}

// Document structure to be stored in MongoDB
type DBDocument struct {
	ID  string      `bson:"_id"`
	Doc interface{} `bson:"Doc"`
}

func DBDocument_new(Doc interface{}) DBDocument {
	return DBDocument{Doc: Doc}
}

// Generic Document structure to be decoded into any type.
type GenericDBDocument struct {
	ID  string          `bson:"_id"`
	Doc GenericDocument `bson:"Doc"`
}

// Generic Document to be decoded or encoded into any type. Equalent to map[string]interface{}
type GenericDocument bson.M

type WriteOperationResponse struct {
	Status int    // 0 = unknown, 1 = success, 2 = failure (Unknown error), others : HTTP status codes (But not used for the HTTP response)
	Action string // Performed action | "insert" | "update" | "dbreq" | "typecast"
	Result string // Targeted ID or error message
}

func (D *GenericDocument) Cast(Template interface{}) interface{} {
	bb, be := bson.Marshal(D)

	if CheckError(be) {
		return nil
	}

	var out interface{}

	err := bson.Unmarshal(bb, &out)

	if CheckError(err) {
		return nil
	}

	return out
}

func ToJson(Obj *interface{}) string {
	jb, je := json.Marshal(*Obj)

	if CheckError(je) {
		return ""
	}
	return string(jb)
}
func ToJsonPretty(Obj *interface{}) string {
	jb, je := json.MarshalIndent(*Obj, "", "    ")

	if CheckError(je) {
		return ""
	}
	return string(jb)
}

type Filter struct {
	MongoQuery bson.D
}

func Filter_MatchAll() *Filter {
	return &Filter{MongoQuery: bson.D{}}
}

func (F *Filter) Equals(filedPath string, value interface{}) *Filter {
	F.MongoQuery = append(F.MongoQuery, bson.E{
		Key:   "Doc." + filedPath,
		Value: value,
	})
	return F
}

func (F *Filter) Exists(filedPath string, exists bool) *Filter {

	var value bson.M

	if exists {
		value = bson.M{"$exists": true}
	} else {
		value = bson.M{"$exists": false}
	}

	F.MongoQuery = append(F.MongoQuery, bson.E{
		Key:   "Doc." + filedPath,
		Value: value,
	})
	return F
}
