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

// MongoDB - Agra Adapter : MonCore
type Moncore struct {
	client *mongo.Client
}

// Default Context with timeout
func DefaultContext() (*context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultContextTimout)

	return &ctx, cancel
}

// Default Long running Context without timeout but with cancel
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

// Disconnect from MongoDB
func (MC *Moncore) Disconnect() error {

	ctx_dbr, cnc_dbr := DefaultContext()
	defer cnc_dbr()

	return MC.client.Disconnect(*ctx_dbr)
}

// Specify Database to use
func (MC *Moncore) Database(name string) *Database {
	db := MC.client.Database(name)
	return &Database{db: db}
}

// MongoDB Database wrapper
type Database struct {
	db *mongo.Database
}

// Specify Collection to use
func (MD *Database) Collection(name string) *Collection {
	return &Collection{MC: MD.db.Collection(name)}
}

// List all collection names in database
func (MD *Database) ListCollectionNames() []string {

	ctx_dbr, cnc_dbr := DefaultContext()
	defer cnc_dbr()

	names, err := MD.db.ListCollectionNames(*ctx_dbr, bson.M{})

	if CheckError(err) {
		return nil
	}
	return names
}

// List all collections in database
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

// MongoDB Collection wrapper
type Collection struct {
	MC *mongo.Collection
}

func (C *Collection) query_curser(filter *Filter) *mongo.Cursor {

	ctx_dbr, cnc_dbr := DefaultContext()
	defer cnc_dbr()

	qcur, qerr := C.MC.Find(*ctx_dbr, filter.MongoQuery)

	if CheckError(qerr) {
		return nil
	}

	return qcur
}

// Query collection with filter.
// You can use Filter_MatchAll() to match all documents and add filters to filter out documents.
// Returns nil if error.
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

// Query collection with filter.
// You can use Filter_MatchAll() to match all documents and add filters to filter out documents.
//
// bufferSize: 0 means unbuffered, which will load all documents at once to memory.
//
// Returns a channel that will be closed when all documents are read.
// Returns nil if error.
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

// Insert or Update document.
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

// Generic Document structure to be decoded into any type.
type GenericDBDocument struct {
	ID  string          `bson:"_id"`
	Doc GenericDocument `bson:"Doc"`
}

// Generic Document to be decoded or encoded into any type. Equalent to map[string]interface{}
type GenericDocument bson.M

// WriteOperationResponse is returned by Write operations.
type WriteOperationResponse struct {
	Status int    // 0 = unknown, 1 = success, 2 = failure (Unknown error), others : HTTP status codes (But not used for the HTTP response)
	Action string // Performed action | "insert" | "update" | "dbreq" | "typecast"
	Result string // Targeted ID or error message
}

// Cast a GenericDocument into Template type
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

// Serialize object into JSON
func ToJson(Obj *interface{}) string {
	jb, je := json.Marshal(*Obj)

	if CheckError(je) {
		return ""
	}
	return string(jb)
}

// Serialize object into JSON prettified (indented)
func ToJsonPretty(Obj *interface{}) string {
	jb, je := json.MarshalIndent(*Obj, "", "    ")

	if CheckError(je) {
		return ""
	}
	return string(jb)
}

// Filter structure to be used with Query() . This contains the matching criteria and the filed
type Filter struct {
	MongoQuery bson.D
}

// Filterlet contains only the matching criteria. This is used to create a Filter
type Filterlet struct {
	Querylet bson.D
}

// Empty filter that matches all
func Filter_MatchAll() *Filter {
	return &Filter{MongoQuery: bson.D{}}
}

// Empty filter that matches all
func Filterlet_new() *Filterlet {
	return &Filterlet{Querylet: bson.D{}}
}

// Create a filterlet that matches if field is equal.
//
// The returned Filterlet and input Filterlet are the same.
func (F *Filterlet) Equals(value interface{}) *Filterlet {
	F.Querylet = append(F.Querylet, bson.E{
		Key:   "$eq",
		Value: value,
	})

	return F
}

// Create a filterlet that matches if field is not equal.
//
// The returned Filterlet and input Filterlet are the same.
func (F *Filterlet) NotEquals(value interface{}) *Filterlet {
	F.Querylet = append(F.Querylet, bson.E{
		Key:   "$ne",
		Value: value,
	})

	return F
}

// Create a filterlet that matches if field is present or absant according to 'exists' parameter.
//
// The returned Filterlet and input Filterlet are the same.
func (F *Filterlet) Exists(exists bool) *Filterlet {

	F.Querylet = append(F.Querylet, bson.E{
		Key:   "$exists",
		Value: exists,
	})

	return F
}

// Add filterlet to filter. The returned filter and input filter are the same.
func (F *Filter) Add(filedPath string, fl *Filterlet) *Filter {

	F.MongoQuery = append(F.MongoQuery, bson.E{
		Key:   "Doc." + filedPath,
		Value: bson.D(fl.Querylet),
	})
	return F
}
