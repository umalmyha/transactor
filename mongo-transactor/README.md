Let's imagine we have `MyService` where transactional behavior is required and 2 repositories `OrderRepository` and `PaymentRepository` which must create order and payment consistently within single transaction, so if either first or second fail - no data is created at all and transaction is rolled back.  
Firstly, for object which requires transactional behavior you can define property with `transactor` interface, so now `trx` can accept any transactor no matter if it is postgres or mongo or anything else. Please, see original example from [mongo-driver](https://www.mongodb.com/docs/manual/core/transactions/) for running transactions, in short, mongo-driver package injects session data to context, so context is source of truth - are we running within transaction or not.  
Add code to your repositories totally ignoring any transactional behavior. You can use repositories in combination with transactor in `MyService` like shown below:
```go
type MyService struct {
	trx transactor // in this example will be maintained via mongo_transactor.NewMongoDriverTransactor
	orderRps OrderRepository // order repository added
	paymentRps PaymentRepository // payment repository added
}

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
