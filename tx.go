package stellarnet

import "github.com/stellar/go/build"

type Tx struct {
	from           SeedStr
	muts           []build.TransactionMutator
	haveMemo       bool
	haveTimebounds bool
	opCount        int
	err            error
}

// NewBaseTx creates a Tx with the common transaction elements.
func NewBaseTx(from SeedStr, seqnoProvider build.SequenceProvider, baseFee uint64) *Tx {
	if baseFee < build.DefaultBaseFee {
		baseFee = build.DefaultBaseFee
	}
	return &Tx{
		from: from,
		muts: []build.TransactionMutator{
			build.SourceAccount{AddressOrSeed: from.SecureNoLogString()},
			Network(),
			build.AutoSequence{SequenceProvider: seqnoProvider},
			build.BaseFee{Amount: baseFee},
		},
	}
}

// AddPaymentOp adds a payment operation to the transaction.
func (t *Tx) AddPaymentOp(to AddressStr, amount string) {
	if t.err != nil {
		return
	}
	if t.IsFull() {
		t.err = ErrTxOpFull
		return
	}

	t.muts = append(t.muts, build.Payment(
		build.Destination{AddressOrSeed: to.String()},
		build.NativeAmount{Amount: amount},
	))
	t.opCount++
}

// AddCreateAccountOp adds a create_account operation to the transaction.
func (t *Tx) AddCreateAccountOp(to AddressStr, amount string) {
	if t.err != nil {
		return
	}
	if t.IsFull() {
		t.err = ErrTxOpFull
		return
	}

	t.muts = append(t.muts, build.CreateAccount(
		build.Destination{AddressOrSeed: to.String()},
		build.NativeAmount{Amount: amount},
	))
	t.opCount++
}

// AddAccountMergeOp adds an account_merge operation to the transaction.
func (t *Tx) AddAccountMergeOp(to AddressStr) {
	if t.err != nil {
		return
	}
	if t.IsFull() {
		t.err = ErrTxOpFull
		return
	}

	t.muts = append(t.muts, build.AccountMerge(
		build.Destination{AddressOrSeed: to.String()},
	))
	t.opCount++
}

// AddInflationDestinationOp adds a set_options operation for the inflation
// destination to the transaction.
func (t *Tx) AddInflationDestinationOp(to AddressStr) {
	if t.err != nil {
		return
	}
	if t.IsFull() {
		t.err = ErrTxOpFull
		return
	}

	t.muts = append(t.muts, build.SetOptions(build.InflationDest(to.String())))
	t.opCount++
}

// AddMemoText adds a text memo to the transaction.  There can only
// be one memo.
func (t *Tx) AddMemoText(memo string) {
	if t.err != nil {
		return
	}
	if t.haveMemo {
		t.err = ErrMemoExists
		return
	}

	t.muts = append(t.muts, build.MemoText{Value: memo})
	t.haveMemo = true
}

// AddMemoID adds an ID memo to the transaction.  There can only
// be one memo.
func (t *Tx) AddMemoID(id *uint64) {
	if t.err != nil {
		return
	}
	if id == nil {
		return
	}
	if t.haveMemo {
		t.err = ErrMemoExists
		return
	}

	t.muts = append(t.muts, build.MemoID{Value: *id})
	t.haveMemo = true
}

// AddTimebounds adds time bounds to the transaction.
func (t *Tx) AddTimebounds(min, max uint64) {
	t.AddBuiltTimebounds(&build.Timebounds{MinTime: min, MaxTime: max})
}

// AddBuiltTimebounds adds time bounds to the transaction with a *build.Timebounds.
func (t *Tx) AddBuiltTimebounds(bt *build.Timebounds) {
	if t.err != nil {
		return
	}
	if bt == nil {
		return
	}
	if t.haveTimebounds {
		t.err = ErrTimeboundsExist
		return
	}

	t.muts = append(t.muts, *bt)
}

// Sign builds the transaction and signs it.
func (t *Tx) Sign() (SignResult, error) {
	if t.err != nil {
		return SignResult{}, errMap(t.err)
	}
	b, err := t.builder()
	if err != nil {
		return SignResult{}, errMap(err)
	}
	return sign(t.from, b)
}

// IsFull returns true if there are already 100 operations in the transaction.
func (t *Tx) IsFull() bool {
	return t.opCount >= 100
}

func (t *Tx) builder() (*build.TransactionBuilder, error) {
	b, err := build.Transaction(t.muts...)
	if err != nil {
		return nil, errMap(err)
	}
	return b, nil
}
