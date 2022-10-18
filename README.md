# transactor

Repository contains packages which are aimed to simplify transaction management in your logic.

### General idea
Since `context.Context` is usually passed to execute any requests to database, it can be the perfect object to inject transaction into. Package provides functionality for wrapping your "within transaction" logic with transactional behavior.

### How to use
Usage might differ depending on database you use and package as a consequence. In general, all packages have common method `WithinTransaction(ctx context.Context, txFunc func(context.Context) error) error`, so in the code where you need transactional behaviour you can define interface with this single method, for instance:

```go
type transactor interface {
	WithinTransaction(ctx context.Context, txFn func(context.Context) error) error
}

type MyTransactionalLogic struct {
	trx transactor
}

func (lgc *MyTransactionalLogic) RunSmthInTx(ctx context.Context) {
	lgc.trx.WithinTransaction(ctx, func(txCtx) {
		// ...pass context with tx injected in your other methods
        })
}
```
To understand usage for concrete database, please, see corresponding package in repository.