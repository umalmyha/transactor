package mongo_transactor

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDriverTransactor represents mongodb transactor behavior
type MongoDriverTransactor interface {
	WithinTransaction(ctx context.Context, txFn func(context.Context) error) error
	WithinTransactionWithOptions(ctx context.Context, txFn func(context.Context) error, opts ...*options.TransactionOptions) error
}

type mongoDriverTransactor struct {
	client *mongo.Client
}

// NewMongoDriverTransactor builds new MongoDriverTransactor
func NewMongoDriverTransactor(client *mongo.Client) MongoDriverTransactor {
	return &mongoDriverTransactor{client: client}
}

// WithinTransaction runs WithinTransactionWithOptions with default tx options
func (t *mongoDriverTransactor) WithinTransaction(ctx context.Context, txFn func(context.Context) error) error {
	return t.WithinTransactionWithOptions(ctx, txFn)
}

// WithinTransactionWithOptions runs logic within transaction passing context with transaction injected into it specifying options
func (t *mongoDriverTransactor) WithinTransactionWithOptions(ctx context.Context, txFn func(context.Context) error, opts ...*options.TransactionOptions) error {
	session, err := t.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		return nil, txFn(sessCtx)
	}, opts...)

	return err
}
