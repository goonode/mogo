
# What's Mogo?
Mogo is a wrapper for mgo (https://github.com/globalsign/mgo) that adds ODM, hooks, validation and population process, to its raw Mongo functions. Mogo started as a fork of the [bongo](https://github.com/go-bongo/bongo) project and aims to be a re-thinking of the already developed concepts, nearest to the backend mgo driver, but without giving up the simplicity of use. It also adds advanced features such as pagination, population of referenced document which belongs to other collections, and index creation on document fields.

Mogo is tested using GoConvey (https://github.com/smartystreets/goconvey)

(note: if you like this repo and want to collaborate, please don't hesitate more ;-)

<!-- [![Build Status](https://travis-ci.org/goonode/mogo.svg?branch=master)](https://travis-ci.org/goonode/mogo.svg?branch=master)

[![Coverage Status](https://coveralls.io/repos/go-mogo/mogo/badge.svg)](https://coveralls.io/r/go-mogo/mogo) -->

# Usage

## Basic Usage

### Import the Library
`go get github.com/goonode/mogo`

`import "github.com/goonode/mogo"`

And install dependencies:

`cd $GOHOME/src/github.com/goonode/mogo && go get .`

### Connect to a Database

Create a new `mogo.Config` instance:

```go
config := &mogo.Config{
	ConnectionString: "localhost",
	Database:         "mogotest",
}
```

Then just call the `Connect` func passing the config, and make sure to handle any connection errors:

```go
connection, err := mogo.Connect(config)

if err != nil {
	log.Fatal(err)
}
```
`Connect` will create a connection for you and also will store this connection in DBConn global var. 
This global var can be used to access to the Connection object from any place inside the application and it will be used
also from all internal functions when an access to the connection is needed. 

If you need to, you can access the raw `mgo` session with `connection.Session`

### Create a Model

A Model contains all information related to the the interface between a Document and the underlying mgo driver. You need to register a Model (and all Models you want to use in your application) before.

To create a new Model you need to define the document struct, attach the DocumentModel struct to it, and than you need to register it to mogo global registry:

```go
type Bongo struct {
	mogo.DocumentModel `bson:",inline" coll:"mogo-registry-coll"`
	Name          string
	Friends       RefField `ref:"Macao"`
}

type Macao struct {
	mogo.DocumentModel `bson:",inline" coll:"mogo-registry-coll"`
	Name          string
}

mogo.ModelRegistry.Register(Bongo{}, Macao{})

```

`ModelRegistry` is an helper struct and can be used to globally register all models of the application. It will be used internally to store information about the document, that will be used to perform internal magics.


### Create a Document

Any struct can be used as a document as long as it embed the `DocumentModel` struct in a field. 
The `DocumentModel` provided with mogo implements the `Document` interface as well as the `Model`, `NewTracker`, `TimeCreatedTracker` and `TimeModifiedTracker` interfaces (to keep track of new/existing documents and created/modified timestamps). 
The `DocumentModel` must be embedded with `bson:",inline"` tag otherwise you will get nested behavior when the data goes to your database. Also it requires the `coll` or `collection` tag which will be used to assign the model to a mongo collection. 
The `coll` tag can be used only on this field of the struct, and each document can only have one collection. The `idx` or `index` tag can be used to create indexes (the index feature is in development stage and very limited at the moment). 
The syntax for the `idx` tag is `{field1,...},unique,sparse,...`. The field name must follow the bson tag specs.

The recommended way to create a new document model instance is by calling the `NewDoc`, that returns a pointer to a newly created document.

```go
type Person struct {
	mogo.DocumentModel `bson:",inline" coll:"user-coll"`
	FirstName string
	LastName string
	Gender string
}

func main() {
	Person := NewDoc(Person{}).(*Person)
	...
}
```

You can use child structs as well.

```go
type HomeAddress struct {
		Street string
		Suite string
		City string
		State string
		Zip string
}

type Person struct {
	mogo.DocumentModel `bson:",inline" coll:"user-coll"`
	FirstName string
	LastName string
	Gender string
	HomeAddress HomeAddress
}

func main() {
	Person := NewDoc(Person{}).(*Person)
	...
}
```

Indexes can be defined using the `idx` tag on the field you want to create the index for. The syntax for the `idx` tag is 
```go
`idx:{field || field1,field2,...},keyword1,keyword2,...`
```

Supported keywords are `unique, sparse, background and dropdups`.

```go
type HomeAddress struct {
		Street string
		Suite string
		City string
		State string
		Zip string
}

type Person struct {
	mogo.DocumentModel `bson:",inline" coll:"user-coll"`
	FirstName string	`idx:"{firstname},unique"`
	LastName string		`idx:"{lastname},unique"`
	Gender string
	HomeAddress HomeAddress `idx:"{homeaddress.street, homeaddress.city},unique,sparse"`
}

func main() {
	Person := NewDoc(Person{}).(*Person)
	...
}
```

Also composite literal can be used to initialize the document before creating a new instance:

```go
func main() {
	Person := NewDoc(Person{
		FirstName: "MyFirstName",
		LastName: "MyLastName",
		...
	}).(*Person)
	...
}
```

#### Hooks

You can add special methods to your document type that will automatically get called by mogo during certain actions. Currently available hooks are:

* `func (s *DocumentStruct) Validate() []error` (returns a slice of errors - if it is empty then it is assumed that validation succeeded)
* `func (s *DocumentStruct) BeforeSave() error`
* `func (s *DocumentStruct) AfterSave() error`
* `func (s *DocumentStruct) BeforeDelete() error`
* `func (s *DocumentStruct) AfterDelete() error`
* `func (s *DocumentStruct) AfterFind() error`

### Saving or Updating Models

To save a document just call `Save()` helper func passing the instance as parameter, or using the instance method of the created
document instance. Actually the `Save()` func make a call to the underlying mgo.UpsertId() func, so it can be used to perform a document
update, too.

```go
myPerson := NewDoc(Person{}).(*Person)
myPerson.FirstName = "Bingo"
myPerson.LastName = "Mogo"

err := Save(myPerson)
```

or the equivalent form using the `Save()` method of the new instance:

```go
myPerson := NewDoc(Person{}).(*Person)
myPerson.FirstName = "Bingo"
myPerson.LastName = "Mogo"

err := myPerson.Save()
```

Now you'll have a new document in the collection `user-coll` as defined into the Person model. 
If there is an error, you can check if it is a validation error using a type assertion:

```go
if vErr, ok := err.(*Bongo.ValidationError); ok {
	fmt.Println("Validation errors are:", vErr.Errors)
} else {
	fmt.Println("Got a real error:", err.Error())
}
```

### Deleting Documents
There are several ways to delete a document.

#### Remove / RemoveAll helper funcs
Same thing as `Save` - just call `Remove` passing the Document instance or RemoveAll by passing a slice of Documents.
```go
err := Remove(person)
```

This *will* run the `BeforeDelete` and `AfterDelete` hooks, if applicable.

#### RemoveBySelector / RemoveAllBySelector helper funcs
This just delegates to `mgo.Collection.Remove` and `mgo.Collection.RemoveAll`. It will *not* run the `BeforeDelete` and `AfterDelete` hooks. The RemoveAllBySelector accepts a map of selectors for which the key is the interface name of the model and returns a map of `*ChangeInfoWithError` one for each passed interface. 

```go
err := RemoveBySelector(bson.M{"FirstName":"Testy"})
```


### Finding

There are several ways to make a find. Finding methods are glued to the mgo driver so each method can use mgo driver directly (but this way also disable the hooks execution). The Query and Iter objects are defined as extensions of the mgo equivalent ones and for this reason all results are to be accessed using the iterator. 

The define a query the Query object can be used as for example:

```go
conn := getConnection()
defer conn.Session.Close()

ModelRegistry.Register(noHookDocument{}, hookedDocument{})

doc := NewDoc(noHookDocument{}).(*noHookDocument)

iter := doc.Find(nil).Iter()
for iter.Next(doc) {
	count++
}
```

### Populate

It is possible to use a document field to store references to other documents. The document field needs to be of type `RefField` or
`RefFieldSlice` and the `ref` tag needs to be attached to that field.

```go
type Bongo struct {
	mogo.DocumentModel `bson:",inline" coll:"mogo-registry"` // The mogo will be stored in the mogo-registry collection
	Name          string
	Friends       RefFieldSlice `ref:"Macao"` // The field Friends of mogo is a reference to a slice of Macao objects
	BestFriend    RefField      `ref:"Macao"`
}

type Macao struct {
	mogo.DocumentModel `bson:",inline" coll:"mogo-registry"` // The Macao will be stored in the mogo-registry collection
	Name          string
}

mogo.ModelRegistry.Register(Bongo{}, Macao{})
```

The `RefField` accepts a bson id that will be stored in the related field of the document. To load the `RefField` with the referenced object, the `Populate()` method will be used. The `Populate()` works on *loaded* document (i.e. document returned by Find()/Iter() methods). The following example show how to use this feature. 

```go
...

bongo := NewDoc(Bongo{}).(*Bongo)

// All friends of bongo are macaos, and now we will give some friends to bongo
for i := 0; i < 10; i++ {
	macao := NewDoc(Macao{}).(*Macao)
	macao.Name = fmt.Sprintf("Macky%d", i)
	Save(macao)
	bongo.Friends = append(bongo.Friends, &RefField{ID: macao.ID})
}

// But bongo best friend is Polly
macao := NewDoc(Macao{}).(*Macao)
macao.Name = "Polly"
Save(macao)

bongo.BestFriend = RefField{ID: macao.ID}
Save(bongo)

// Now bongo.Friends contains a lot of ids, now we need to access to their data
q := bongo.Populate("Friends").All()

...
```
The `Populate()` method returns a special kind of `mogo.Query` object, for the referenced object, for which it is possible to chain a filter using the `Find()` method. In this case a bson.M type should be used as query interface.

```go
...
q := bongo.Populate("Friends").Find(bson.M{"name": "Macky3"}).All()
...
```


### Pagination: Paginate and NextPage
To enable pagination you need to call the `Paginate()` method and the `NextPage()` iterator.

```go
conn := getConnection()
defer conn.Session.Close()

mogo.ModelRegistry.Register(noHookDocument{}, hookedDocument{})

doc := NewDoc(noHookDocument{}).(*noHookDocument)

iter := doc.Find(nil).Paginate(3).Iter()
results := make([]*noHookDocument, 3) 

for iter.NextPage(&results) {
	...
}

```


### FindOne and FindByID helper funcs
You can use `doc.FindOne()` and `doc.FindByID()` as replacement of `doc.Find().One()` and `doc.FindID().One()` 


## Change Tracking
If your model struct implements the `Trackable` interface, it will automatically track changes to your model so you can compare the current values with the original. For example:

```go
type MyModel struct {
	mogo.DocumentModel `bson:",inline"`
	StringVal string
	diffTracker *Bongo.DiffTracker
}

// Easy way to lazy load a diff tracker
func (m *MyModel) GetDiffTracker() *DiffTracker {
	if m.diffTracker == nil {
		m.diffTracker = mogo.NewDiffTracker(m)
	}

	return m.diffTracker
}

myModel := NewDoc(&MyModel{}).(*MyModel{})
```

Use as follows:

### Check if a field has been modified
```go
// Store the current state for comparison
myModel.GetDiffTracker().Reset()

// Change a property...
myModel.StringVal = "foo"

// We know it's been instantiated so no need to use GetDiffTracker()
fmt.Println(myModel.diffTracker.Modified("StringVal")) // true
myModel.diffTracker.Reset()
fmt.Println(myModel.diffTracker.Modified("StringVal")) // false
```

### Get all modified fields
```go
myModel.StringVal = "foo"
// Store the current state for comparison
myModel.GetDiffTracker().Reset()

isNew, modifiedFields := myModel.GetModified()

fmt.Println(isNew, modifiedFields) // false, ["StringVal"]
myModel.diffTracker.Reset()

isNew, modifiedFields = myModel.GetModified()
fmt.Println(isNew, modifiedFields) // false, []
```

### Diff-tracking Session
If you are going to be checking more than one field, you should instantiate a new `DiffTrackingSession` with `diffTracker.NewSession(useBsonTags bool)`. This will load the changed fields into the session. Otherwise with each call to `diffTracker.Modified()`, it will have to recalculate the changed fields.
