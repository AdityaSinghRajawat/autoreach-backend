package unit

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type enumValue string

type columnsFixture struct {
	ID      string `db:"id"`
	Name    string `db:"name"`
	Ignored string `db:"-"`
	Plain   string
}

type fakeRow struct{ err error }

func (r fakeRow) Scan(...any) error { return r.err }

func TestBuildKeySkipsEmptyPartsAndUsesTheDelimiter(t *testing.T) {
	assert.Equal(t, "a:b:c", utils.BuildKey(":", "a", "b", "c"))
	assert.Equal(t, "a:c", utils.BuildKey(":", "a", "", "c"), "no doubled delimiter")
	assert.Equal(t, "a-b", utils.BuildKey("-", "", "a", "b", ""))
	assert.Empty(t, utils.BuildKey(":"))
}

func TestGenerateUUIDReturnsDistinctValidUUIDs(t *testing.T) {
	first, second := utils.GenerateUUID(), utils.GenerateUUID()

	require.NoError(t, uuid.Validate(first))
	assert.NotEqual(t, first, second)
}

func TestSelectColumnsFollowTheDBTagsAndSkipUntaggedFields(t *testing.T) {
	assert.Equal(t, "c.id, c.name", utils.BuildSelectColumns[columnsFixture]("c"))
}

func TestSQLEnumListQuotesAndEscapesLiterals(t *testing.T) {
	assert.Equal(t, "'a', 'b'", utils.BuildSQLEnumList(enumValue("a"), enumValue("b")))
	assert.Equal(t, "'it''s'", utils.BuildSQLEnumList("it's"))
	assert.Empty(t, utils.BuildSQLEnumList[enumValue]())
}

func TestSortedIndexesByKeyIsStableAndLeavesItemsInPlace(t *testing.T) {
	items := []string{"b1", "a1", "b2", "a2"}

	order := utils.SortedIndexesByKey(items, func(item string) string { return item[:1] })

	assert.Equal(t, []int{1, 3, 0, 2}, order)
	assert.Equal(t, []string{"b1", "a1", "b2", "a2"}, items)
	assert.Empty(
		t,
		utils.SortedIndexesByKey([]string(nil), func(item string) string { return item }),
	)
}

func TestScanOptionalRowTurnsAMissingRowIntoNilAndWrapsOtherErrors(t *testing.T) {
	found := func(utils.RowScanner) (*string, error) {
		value := "ok"

		return &value, nil
	}
	record, err := utils.ScanOptionalRow(fakeRow{}, found, "find")
	require.NoError(t, err)
	assert.Equal(t, "ok", *record)

	boom := errors.New("boom")
	failing := func(utils.RowScanner) (*string, error) { return nil, boom }
	_, err = utils.ScanOptionalRow(fakeRow{}, failing, "find thing")
	require.ErrorIs(t, err, boom)
	assert.Contains(t, err.Error(), "find thing")
}

func TestOptionalStringHelpersKeepNilAndNormalizeText(t *testing.T) {
	padded, blank := "  Go Dev  ", "   "

	assert.Nil(t, utils.TrimOptionalString(nil))
	assert.Equal(t, "Go Dev", *utils.TrimOptionalString(&padded))
	assert.Equal(t, "", *utils.TrimOptionalString(&blank))

	assert.Nil(t, utils.TrimLowerOptionalString(nil))
	assert.Equal(t, "go dev", *utils.TrimLowerOptionalString(&padded))

	assert.Nil(t, utils.TrimOptionalStringToNil(nil))
	assert.Nil(t, utils.TrimOptionalStringToNil(&blank))
	assert.Equal(t, "Go Dev", *utils.TrimOptionalStringToNil(&padded))
}

func TestFullNameIsSplitIntoFirstAndRemainingWords(t *testing.T) {
	first, rest := utils.SplitFullName("  Alex  van  Dyke ")
	assert.Equal(t, "Alex", first)
	assert.Equal(t, "van Dyke", rest)

	first, rest = utils.SplitFullName("Alex")
	assert.Equal(t, "Alex", first)
	assert.Empty(t, rest)

	first, rest = utils.SplitFullName("   ")
	assert.Empty(t, first)
	assert.Empty(t, rest)

	assert.Equal(t, "Alex", utils.GetFirstName("Alex Candidate"))
	assert.Empty(t, utils.GetFirstName(""))
}

func TestJoinWithAndReadsAsANaturalList(t *testing.T) {
	assert.Empty(t, utils.JoinWithAnd(nil))
	assert.Equal(t, "Go", utils.JoinWithAnd([]string{"Go"}))
	assert.Equal(t, "Go and SQL", utils.JoinWithAnd([]string{"Go", "SQL"}))
	assert.Equal(t, "Go, SQL and Redis", utils.JoinWithAnd([]string{"Go", "SQL", "Redis"}))
}

func TestTruncateRunesCutsOnCharactersNotBytes(t *testing.T) {
	assert.Equal(t, "héл", utils.TruncateRunes("héллo", 3))
	assert.Equal(t, "short", utils.TruncateRunes("short", 50))
	assert.Empty(t, utils.TruncateRunes("anything", 0))
}

func TestUniqueTrimmedValuesDropsBlanksAndRepeatsAndCapsTheResult(t *testing.T) {
	values := []string{" Go ", "go", "", "SQL", "Go", "Redis"}

	assert.Equal(t, []string{"Go", "go", "SQL"}, utils.UniqueTrimmedValues(values, nil, 3))
	assert.Equal(
		t,
		[]string{"go", "sql", "redis"},
		utils.UniqueTrimmedValues(values, strings.ToLower, 10),
	)
}

func TestFirstOrZeroReturnsTheFirstItemOrTheZeroValue(t *testing.T) {
	assert.Equal(t, "a", utils.FirstOrZero([]string{"a", "b"}))
	assert.Empty(t, utils.FirstOrZero([]string(nil)))
	assert.Zero(t, utils.FirstOrZero([]int{}))
}

func TestIsHTTPURLAcceptsOnlyWebLinksWithAHost(t *testing.T) {
	assert.True(t, utils.IsHTTPURL("https://example.com/a?b=1"))
	assert.True(t, utils.IsHTTPURL("http://example.com"))

	for _, link := range []string{"", "ftp://example.com", "javascript:alert(1)", "https://", "example.com", "mailto:a@b.co"} {
		assert.False(t, utils.IsHTTPURL(link), "link %q", link)
	}
}
