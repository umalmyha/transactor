Let's imagine we have `MyService` where transactional behavior is required and 2 repositories `OrderRepository` and `PaymentRepository` which must create order and payment consistently within single transaction, so if either first or second fail - no data is created at all and transaction is rolled back.

Firstly, for object which requires transactional behavior you can define property with `transactor` interface, so now `trx` can accept any transactor no matter if it is postgres or mongo or anything else:
```go
type transactor interface {
	WithinTransaction(ctx context.Context, txFn func(context.Context) error) error
}

type MyService struct {
	trx transactor
}
```
Then, your database accessors (repositories in our case) must be embedded with concrete transaction runner interface, in current example - `PgxWithinTransactionRunner`. To run query you call `Runner` method of transaction runner (it is available because of embedding) and run query. Please, pay attention on `OrderRepository` method `CreateOrder` below:
```go
type Order struct {
	// Order fields
}

type Payment struct {
	// Payment fields
}

type OrderRepository struct {
	pgx_transactor.PgxWithinTransactionRunner
}

func NewOrderRepository(r pgx_transactor.PgxWithinTransactionRunner) *OrderRepository {
	return &OrderRepository{PgxWithinTransactionRunner: r}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *Order) error {
	// queryRunner can be used now to perform any pgx related functionality available in interface `PgxQueryRunner` 
	// it returns either `pgx.Tx` or `pgxpool.Pool` both implements `PgxQueryRunner` interface
	queryRunner := r.Runner(ctx)
	// ...
	
	// so, example of running INSERT query
	q := "INSERT ..."
	args := ...
	r.Runner(ctx).Exec(ctx, q, args...)
}

type PaymentRepository struct {
	pgx_transactor.PgxWithinTransactionRunner
}

func NewPaymentRepository(r pgx_transactor.PgxWithinTransactionRunner) *PaymentRepository {
	return &PaymentRepository{PgxWithinTransactionRunner: r}
}

func (r *PaymentRepository) CreatePayment(ctx context.Context, p *model.Payment) error { 
	// pretty the same concept as for CreateOrder of OrderRepository 
	//...
}
```
Now, you can use repositories in combination with transactor in `MyService`:
```go
type MyService struct {
	trx transactor // in this example will be maintained via pgx_transactor.NewPgxTransactor(...)
	orderRps OrderRepository // order repository added
	paymentRps PaymentRepository // payment repository added
}

// ...
// omitting NewMyService
// ...

// new method CreateOrderWithItems is added
func (s *MyService) CreateOrderAndPayment(ctx context.Context, order *Order, payment *Payment) error {
	// WithinTransaction injects pgx.Tx into ctx, so txCtx is already with injected transaction
	err := s.trx.WithinTransaction(ctx, func (txCtx context.Context) error {
		var inTxErr error

		// txCtx has pgx.Tx injected - Runner in repository will return pgx.Tx
		inTxErr = s.orderRps.CreateOrder(txCtx, order)
		if inTxErr != nil {
			// if error - transaction is rolled back
			return inTxErr
		}
		
		// txCtx has pgx.Tx injected - Runner in repository will return pgx.Tx
		inTxErr = s.paymentRps.CreatePayment(txCtx, payment)
		if inTxErr != nil {
			// if error - transaction is rolled back 
			return inTxErr
		}
		
		// transaction is committed if returned error is nil
		return nil
	})
}
```
As you can see from example above, transaction is shared between 2 repositories through context where it is injected
