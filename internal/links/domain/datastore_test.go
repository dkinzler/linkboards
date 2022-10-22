package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkQueryParams(t *testing.T) {
	a := assert.New(t)

	qp := NewLinkQueryParams()
	a.Equal(20, qp.Limit)
	a.EqualValues(SortOrderNewest, qp.SortOrder)

	qp = qp.SortByTop()
	a.EqualValues(SortOrderTop, qp.SortOrder)

	qp = qp.SortByNewest()
	a.EqualValues(SortOrderNewest, qp.SortOrder)

	qp = qp.WithLimit(50)
	a.Equal(50, qp.Limit)

	// out of range limit
	qp = qp.WithLimit(1000)
	a.Equal(50, qp.Limit)

	a.Nil(qp.CursorScore)

	qp = qp.WithScoreCursor(20)
	a.Equal(20, *qp.CursorScore)

	qp = qp.WithCreatedTimeCursor(133742)
	a.EqualValues(133742, qp.CursorCreatedTime)
}
