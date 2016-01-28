package hooktest

import (
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
	"golang.org/x/net/context"
)

type StackingHook struct {
	Context         []context.Context
	Records         []*skydb.Record
	OriginalRecords []*skydb.Record
}

func (p *StackingHook) Func(ctx context.Context, record *skydb.Record, originalRecord *skydb.Record) skyerr.Error {
	p.Context = append(p.Context, ctx)
	p.Records = append(p.Records, record)
	p.OriginalRecords = append(p.OriginalRecords, originalRecord)
	return nil
}
