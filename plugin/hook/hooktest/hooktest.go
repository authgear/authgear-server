package hooktest

import (
	"github.com/oursky/skygear/skydb"
	"golang.org/x/net/context"
)

type StackingHook struct {
	Context         []context.Context
	Records         []*skydb.Record
	OriginalRecords []*skydb.Record
}

func (p *StackingHook) Func(ctx context.Context, record *skydb.Record, originalRecord *skydb.Record) error {
	p.Context = append(p.Context, ctx)
	p.Records = append(p.Records, record)
	p.OriginalRecords = append(p.OriginalRecords, originalRecord)
	return nil
}
