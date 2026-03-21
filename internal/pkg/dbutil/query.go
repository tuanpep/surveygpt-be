package dbutil

import (
	"strconv"
	"time"
)

// QueryBuilder helps construct dynamic WHERE clauses and argument lists
// for SQL queries with proper parameterized placeholders.
type QueryBuilder struct {
	where string
	args  []any
	argN  int
}

// NewQuery creates a QueryBuilder with an initial WHERE clause and first argument.
// argN should be set to the next placeholder number (usually 2 if $1 is already used).
func NewQuery(baseWhere string, firstArg any, argN int) *QueryBuilder {
	return &QueryBuilder{
		where: baseWhere,
		args:  []any{firstArg},
		argN:  argN,
	}
}

// AndEqual adds a col = $N condition.
func (q *QueryBuilder) AndEqual(col string, val any) *QueryBuilder {
	q.where += " AND " + col + " = $" + strconv.Itoa(q.argN)
	q.args = append(q.args, val)
	q.argN++
	return q
}

// AndLike adds a col ILIKE $N condition with %val% pattern.
func (q *QueryBuilder) AndLike(col string, val string) *QueryBuilder {
	q.where += " AND " + col + " ILIKE $" + strconv.Itoa(q.argN)
	q.args = append(q.args, "%"+val+"%")
	q.argN++
	return q
}

// AndGreaterEqual adds a col >= $N condition.
func (q *QueryBuilder) AndGreaterEqual(col string, val any) *QueryBuilder {
	q.where += " AND " + col + " >= $" + strconv.Itoa(q.argN)
	q.args = append(q.args, val)
	q.argN++
	return q
}

// AndLessEqual adds a col <= $N condition.
func (q *QueryBuilder) AndLessEqual(col string, val any) *QueryBuilder {
	q.where += " AND " + col + " <= $" + strconv.Itoa(q.argN)
	q.args = append(q.args, val)
	q.argN++
	return q
}

// AndLessThan adds a col < $N condition.
func (q *QueryBuilder) AndLessThan(col string, val any) *QueryBuilder {
	q.where += " AND " + col + " < $" + strconv.Itoa(q.argN)
	q.args = append(q.args, val)
	q.argN++
	return q
}

// AndCursor adds a time-based cursor condition: col < $N.
func (q *QueryBuilder) AndCursor(col string, t time.Time) *QueryBuilder {
	q.where += " AND " + col + " < $" + strconv.Itoa(q.argN)
	q.args = append(q.args, t)
	q.argN++
	return q
}

// Where returns the constructed WHERE clause string.
func (q *QueryBuilder) Where() string {
	return q.where
}

// Args returns all accumulated arguments.
func (q *QueryBuilder) Args() []any {
	return q.args
}

// ArgsWithoutLast returns all arguments except the last one (useful when the last arg is LIMIT).
func (q *QueryBuilder) ArgsWithoutLast() []any {
	if len(q.args) <= 1 {
		return q.args
	}
	return q.args[:len(q.args)-1]
}

// NextArg returns the current argument index and increments it.
func (q *QueryBuilder) NextArg() int {
	n := q.argN
	q.argN++
	return n
}

// AppendLimit adds a LIMIT $N argument and returns the placeholder string.
func (q *QueryBuilder) AppendLimit(limit int) string {
	placeholder := "$" + strconv.Itoa(q.argN)
	q.args = append(q.args, limit)
	q.argN++
	return placeholder
}
