package mongodb

import (
	"context"
	"log"
	"time"

	"github.com/nicholasss/expense-tracker-api/internal/expenses"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// database type(s)

// mongoExpense has time stored as unix seconds (not milli-)
type mongoExpense struct {
	ID          int
	CreatedAt   int64
	OccuredAt   int64
	Description string
	Amount      int64
}

func toMongoExpense(e *expenses.Expense) mongoExpense {
	// convert times to int
	return mongoExpense{
		ID:          e.ID,
		Description: e.Description,
		Amount:      e.Amount,
		// CreatedAt will occur within the database
		OccuredAt: e.ExpenseOccuredAt.Unix(),
	}
}

func toServiceExpense(db mongoExpense) *expenses.Expense {
	return &expenses.Expense{
		ID:               db.ID,
		Description:      db.Description,
		Amount:           db.Amount,
		RecordCreatedAt:  time.Unix(db.CreatedAt, 0),
		ExpenseOccuredAt: time.Unix(db.OccuredAt, 0),
	}
}

// repository type & constructor

type MongoDBRespository struct {
	Client *mongo.Client
}

func NewMongoDBRespository(uri string) (*MongoDBRespository, error) {
	if uri == "" {
		log.Fatal("MongoDB string is empty. Please check config and .env")
	}

	opts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}

	log.Printf("Successfully connected to mongodb at: %v", uri)

	return &MongoDBRespository{Client: client}, nil
}

// implementation of/conformance to interface

func (r *MongoDBRespository) GetByID(ctx context.Context, id int) (*expenses.Expense, error) {
	coll := r.Client.Database("expenses-api").Collection("expenses")

	var record mongoExpense
	result := coll.FindOne(ctx, bson.D{{Key: "expense-id", Value: id}})
	err := result.Decode(&record)
	if err != nil {
		log.Printf("error from GetByID(): %v", err)
		return nil, expenses.ErrUnusedID
	}

	return toServiceExpense(record), nil
}

func (r *MongoDBRespository) GetAll(ctx context.Context) ([]*expenses.Expense, error) {
	log.Print("MongoDBRepository.GetAll() not yet implmeneted!")
	return nil, nil
}

func (r *MongoDBRespository) Create(ctx context.Context, exp *expenses.Expense) (*expenses.Expense, error) {
	coll := r.Client.Database("expenses-api").Collection("expenses")

	record := toMongoExpense(exp)
	result, err := coll.InsertOne(ctx, record)
	if err != nil {
		log.Printf("error from Create(): %v", err)
		return nil, err
	}

	log.Printf("inserted id: %v", result.InsertedID)

	return exp, nil
}

func (r *MongoDBRespository) Update(ctx context.Context, exp *expenses.Expense) error {
	log.Print("MongoDBRepository.Update() not yet implmeneted!")
	return nil
}

func (r *MongoDBRespository) Delete(ctx context.Context, id int) error {
	log.Print("MongoDBRepository.Delete() not yet implmeneted!")
	return nil
}
