package mibs

type EntrySyntax []*Object

type Table struct {
	ID

	IndexSyntax IndexSyntax
	EntrySyntax EntrySyntax
}
