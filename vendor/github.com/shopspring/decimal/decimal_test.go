package decimal

import (
	"encoding/json"
	"encoding/xml"
	"math"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

var testTable = map[float64]string{
	3.141592653589793:   "3.141592653589793",
	3:                   "3",
	1234567890123456:    "1234567890123456",
	1234567890123456000: "1234567890123456000",
	1234.567890123456:   "1234.567890123456",
	.1234567890123456:   "0.1234567890123456",
	0:                   "0",
	.1111111111111110:   "0.111111111111111",
	.1111111111111111:   "0.1111111111111111",
	.1111111111111119:   "0.1111111111111119",
	.000000000000000001: "0.000000000000000001",
	.000000000000000002: "0.000000000000000002",
	.000000000000000003: "0.000000000000000003",
	.000000000000000005: "0.000000000000000005",
	.000000000000000008: "0.000000000000000008",
	.1000000000000001:   "0.1000000000000001",
	.1000000000000002:   "0.1000000000000002",
	.1000000000000003:   "0.1000000000000003",
	.1000000000000005:   "0.1000000000000005",
	.1000000000000008:   "0.1000000000000008",
	1e25:                "10000000000000000000000000",
}

var testTableScientificNotation = map[string]string{
	"1e9":        "1000000000",
	"2.41E-3":    "0.00241",
	"24.2E-4":    "0.00242",
	"243E-5":     "0.00243",
	"1e-5":       "0.00001",
	"245E3":      "245000",
	"1.2345E-1":  "0.12345",
	"0e5":        "0",
	"0e-5":       "0",
	"123.456e0":  "123.456",
	"123.456e2":  "12345.6",
	"123.456e10": "1234560000000",
}

func init() {
	// add negatives
	for f, s := range testTable {
		if f > 0 {
			testTable[-f] = "-" + s
		}
	}
	for e, s := range testTableScientificNotation {
		if string(e[0]) != "-" && s != "0" {
			testTableScientificNotation["-"+e] = "-" + s
		}
	}
}

func TestNewFromFloat(t *testing.T) {
	for f, s := range testTable {
		d := NewFromFloat(f)
		if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.value.String(), d.exp)
		}
	}

	shouldPanicOn := []float64{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
	}

	for _, n := range shouldPanicOn {
		var d Decimal
		if !didPanic(func() { d = NewFromFloat(n) }) {
			t.Fatalf("Expected panic when creating a Decimal from %v, got %v instead", n, d.String())
		}
	}
}

func TestNewFromString(t *testing.T) {
	for _, s := range testTable {
		d, err := NewFromString(s)
		if err != nil {
			t.Errorf("error while parsing %s", s)
		} else if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.value.String(), d.exp)
		}
	}

	for e, s := range testTableScientificNotation {
		d, err := NewFromString(e)
		if err != nil {
			t.Errorf("error while parsing %s", e)
		} else if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.value.String(), d.exp)
		}
	}
}

func TestNewFromStringErrs(t *testing.T) {
	tests := []string{
		"",
		"qwert",
		"-",
		".",
		"-.",
		".-",
		"234-.56",
		"234-56",
		"2-",
		"..",
		"2..",
		"..2",
		".5.2",
		"8..2",
		"8.1.",
		"1e",
		"1-e",
		"1e9e",
		"1ee9",
		"1ee",
		"1eE",
		"1e-",
		"1e-.",
		"1e1.2",
		"123.456e1.3",
		"1e-1.2",
		"123.456e-1.3",
		"123.456Easdf",
		"123.456e" + strconv.FormatInt(math.MinInt64, 10),
		"123.456e" + strconv.FormatInt(math.MinInt32, 10),
	}

	for _, s := range tests {
		_, err := NewFromString(s)

		if err == nil {
			t.Errorf("error expected when parsing %s", s)
		}
	}
}

func TestNewFromFloatWithExponent(t *testing.T) {
	type Inp struct {
		float float64
		exp   int32
	}
	tests := map[Inp]string{
		Inp{123.4, -3}:      "123.4",
		Inp{123.4, -1}:      "123.4",
		Inp{123.412345, 1}:  "120",
		Inp{123.412345, 0}:  "123",
		Inp{123.412345, -5}: "123.41235",
		Inp{123.412345, -6}: "123.412345",
		Inp{123.412345, -7}: "123.412345",
	}

	// add negatives
	for p, s := range tests {
		if p.float > 0 {
			tests[Inp{-p.float, p.exp}] = "-" + s
		}
	}

	for input, s := range tests {
		d := NewFromFloatWithExponent(input.float, input.exp)
		if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.value.String(), d.exp)
		}
	}

	shouldPanicOn := []float64{
		math.NaN(),
		math.Inf(1),
		math.Inf(-1),
	}

	for _, n := range shouldPanicOn {
		var d Decimal
		if !didPanic(func() { d = NewFromFloatWithExponent(n, 0) }) {
			t.Fatalf("Expected panic when creating a Decimal from %v, got %v instead", n, d.String())
		}
	}
}

func TestJSON(t *testing.T) {
	for _, s := range testTable {
		var doc struct {
			Amount Decimal `json:"amount"`
		}
		docStr := `{"amount":"` + s + `"}`
		err := json.Unmarshal([]byte(docStr), &doc)
		if err != nil {
			t.Errorf("error unmarshaling %s: %v", docStr, err)
		} else if doc.Amount.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, doc.Amount.String(),
				doc.Amount.value.String(), doc.Amount.exp)
		}

		out, err := json.Marshal(&doc)
		if err != nil {
			t.Errorf("error marshaling %+v: %v", doc, err)
		} else if string(out) != docStr {
			t.Errorf("expected %s, got %s", docStr, string(out))
		}
	}
}

func TestBadJSON(t *testing.T) {
	for _, testCase := range []string{
		"]o_o[",
		"{",
		`{"amount":""`,
		`{"amount":""}`,
		`{"amount":"nope"}`,
		`0.333`,
	} {
		var doc struct {
			Amount Decimal `json:"amount"`
		}
		err := json.Unmarshal([]byte(testCase), &doc)
		if err == nil {
			t.Errorf("expected error, got %+v", doc)
		}
	}
}

func TestXML(t *testing.T) {
	for _, s := range testTable {
		var doc struct {
			XMLName xml.Name `xml:"account"`
			Amount  Decimal  `xml:"amount"`
		}
		docStr := `<account><amount>` + s + `</amount></account>`
		err := xml.Unmarshal([]byte(docStr), &doc)
		if err != nil {
			t.Errorf("error unmarshaling %s: %v", docStr, err)
		} else if doc.Amount.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, doc.Amount.String(),
				doc.Amount.value.String(), doc.Amount.exp)
		}

		out, err := xml.Marshal(&doc)
		if err != nil {
			t.Errorf("error marshaling %+v: %v", doc, err)
		} else if string(out) != docStr {
			t.Errorf("expected %s, got %s", docStr, string(out))
		}
	}
}

func TestBadXML(t *testing.T) {
	for _, testCase := range []string{
		"o_o",
		"<abc",
		"<account><amount>7",
		`<html><body></body></html>`,
		`<account><amount></amount></account>`,
		`<account><amount>nope</amount></account>`,
		`0.333`,
	} {
		var doc struct {
			XMLName xml.Name `xml:"account"`
			Amount  Decimal  `xml:"amount"`
		}
		err := xml.Unmarshal([]byte(testCase), &doc)
		if err == nil {
			t.Errorf("expected error, got %+v", doc)
		}
	}
}

func TestDecimal_rescale(t *testing.T) {
	type Inp struct {
		int     int64
		exp     int32
		rescale int32
	}
	tests := map[Inp]string{
		Inp{1234, -3, -5}: "1.234",
		Inp{1234, -3, 0}:  "1",
		Inp{1234, 3, 0}:   "1234000",
		Inp{1234, -4, -4}: "0.1234",
	}

	// add negatives
	for p, s := range tests {
		if p.int > 0 {
			tests[Inp{-p.int, p.exp, p.rescale}] = "-" + s
		}
	}

	for input, s := range tests {
		d := New(input.int, input.exp).rescale(input.rescale)

		if d.String() != s {
			t.Errorf("expected %s, got %s (%s, %d)",
				s, d.String(),
				d.value.String(), d.exp)
		}

		// test StringScaled
		s2 := New(input.int, input.exp).StringScaled(input.rescale)
		if s2 != s {
			t.Errorf("expected %s, got %s", s, s2)
		}
	}
}

func TestDecimal_Floor(t *testing.T) {
	type testData struct {
		input    string
		expected string
	}
	tests := []testData{
		{"1.999", "1"},
		{"1", "1"},
		{"1.01", "1"},
		{"0", "0"},
		{"0.9", "0"},
		{"0.1", "0"},
		{"-0.9", "-1"},
		{"-0.1", "-1"},
		{"-1.00", "-1"},
		{"-1.01", "-2"},
		{"-1.999", "-2"},
	}
	for _, test := range tests {
		d, _ := NewFromString(test.input)
		expected, _ := NewFromString(test.expected)
		got := d.Floor()
		if !got.Equals(expected) {
			t.Errorf("Floor(%s): got %s, expected %s", d, got, expected)
		}
	}
}

func TestDecimal_Ceil(t *testing.T) {
	type testData struct {
		input    string
		expected string
	}
	tests := []testData{
		{"1.999", "2"},
		{"1", "1"},
		{"1.01", "2"},
		{"0", "0"},
		{"0.9", "1"},
		{"0.1", "1"},
		{"-0.9", "0"},
		{"-0.1", "0"},
		{"-1.00", "-1"},
		{"-1.01", "-1"},
		{"-1.999", "-1"},
	}
	for _, test := range tests {
		d, _ := NewFromString(test.input)
		expected, _ := NewFromString(test.expected)
		got := d.Ceil()
		if !got.Equals(expected) {
			t.Errorf("Ceil(%s): got %s, expected %s", d, got, expected)
		}
	}
}

func TestDecimal_RoundAndStringFixed(t *testing.T) {
	type testData struct {
		input         string
		places        int32
		expected      string
		expectedFixed string
	}
	tests := []testData{
		{"1.454", 0, "1", ""},
		{"1.454", 1, "1.5", ""},
		{"1.454", 2, "1.45", ""},
		{"1.454", 3, "1.454", ""},
		{"1.454", 4, "1.454", "1.4540"},
		{"1.454", 5, "1.454", "1.45400"},
		{"1.554", 0, "2", ""},
		{"1.554", 1, "1.6", ""},
		{"1.554", 2, "1.55", ""},
		{"0.554", 0, "1", ""},
		{"0.454", 0, "0", ""},
		{"0.454", 5, "0.454", "0.45400"},
		{"0", 0, "0", ""},
		{"0", 1, "0", "0.0"},
		{"0", 2, "0", "0.00"},
		{"0", -1, "0", ""},
		{"5", 2, "5", "5.00"},
		{"5", 1, "5", "5.0"},
		{"5", 0, "5", ""},
		{"500", 2, "500", "500.00"},
		{"545", -1, "550", ""},
		{"545", -2, "500", ""},
		{"545", -3, "1000", ""},
		{"545", -4, "0", ""},
		{"499", -3, "0", ""},
		{"499", -4, "0", ""},
	}

	// add negative number tests
	for _, test := range tests {
		expected := test.expected
		if expected != "0" {
			expected = "-" + expected
		}
		expectedStr := test.expectedFixed
		if strings.ContainsAny(expectedStr, "123456789") && expectedStr != "" {
			expectedStr = "-" + expectedStr
		}
		tests = append(tests,
			testData{"-" + test.input, test.places, expected, expectedStr})
	}

	for _, test := range tests {
		d, err := NewFromString(test.input)
		if err != nil {
			panic(err)
		}

		// test Round
		expected, err := NewFromString(test.expected)
		if err != nil {
			panic(err)
		}
		got := d.Round(test.places)
		if !got.Equals(expected) {
			t.Errorf("Rounding %s to %d places, got %s, expected %s",
				d, test.places, got, expected)
		}

		// test StringFixed
		if test.expectedFixed == "" {
			test.expectedFixed = test.expected
		}
		gotStr := d.StringFixed(test.places)
		if gotStr != test.expectedFixed {
			t.Errorf("(%s).StringFixed(%d): got %s, expected %s",
				d, test.places, gotStr, test.expectedFixed)
		}
	}
}

func TestDecimal_Uninitialized(t *testing.T) {
	a := Decimal{}
	b := Decimal{}

	decs := []Decimal{
		a,
		a.rescale(10),
		a.Abs(),
		a.Add(b),
		a.Sub(b),
		a.Mul(b),
		a.Div(New(1, -1)),
		a.Round(2),
		a.Floor(),
		a.Ceil(),
		a.Truncate(2),
	}

	for _, d := range decs {
		if d.String() != "0" {
			t.Errorf("expected 0, got %s", d.String())
		}
		if d.StringFixed(3) != "0.000" {
			t.Errorf("expected 0, got %s", d.StringFixed(3))
		}
		if d.StringScaled(-2) != "0" {
			t.Errorf("expected 0, got %s", d.StringScaled(-2))
		}
	}

	if a.Cmp(b) != 0 {
		t.Errorf("a != b")
	}
	if a.Exponent() != 0 {
		t.Errorf("a.Exponent() != 0")
	}
	if a.IntPart() != 0 {
		t.Errorf("a.IntPar() != 0")
	}
	f, _ := a.Float64()
	if f != 0 {
		t.Errorf("a.Float64() != 0")
	}
	if a.Rat().RatString() != "0" {
		t.Errorf("a.Rat() != 0, got %s", a.Rat().RatString())
	}
}

func TestDecimal_Add(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"2", "3"}:                     "5",
		Inp{"2454495034", "3451204593"}:   "5905699627",
		Inp{"24544.95034", ".3451204593"}: "24545.2954604593",
		Inp{".1", ".1"}:                   "0.2",
		Inp{".1", "-.1"}:                  "0",
		Inp{"0", "1.001"}:                 "1.001",
	}

	for inp, res := range inputs {
		a, err := NewFromString(inp.a)
		if err != nil {
			t.FailNow()
		}
		b, err := NewFromString(inp.b)
		if err != nil {
			t.FailNow()
		}
		c := a.Add(b)
		if c.String() != res {
			t.Errorf("expected %s, got %s", res, c.String())
		}
	}
}

func TestDecimal_Sub(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"2", "3"}:                     "-1",
		Inp{"12", "3"}:                    "9",
		Inp{"-2", "9"}:                    "-11",
		Inp{"2454495034", "3451204593"}:   "-996709559",
		Inp{"24544.95034", ".3451204593"}: "24544.6052195407",
		Inp{".1", "-.1"}:                  "0.2",
		Inp{".1", ".1"}:                   "0",
		Inp{"0", "1.001"}:                 "-1.001",
		Inp{"1.001", "0"}:                 "1.001",
		Inp{"2.3", ".3"}:                  "2",
	}

	for inp, res := range inputs {
		a, err := NewFromString(inp.a)
		if err != nil {
			t.FailNow()
		}
		b, err := NewFromString(inp.b)
		if err != nil {
			t.FailNow()
		}
		c := a.Sub(b)
		if c.String() != res {
			t.Errorf("expected %s, got %s", res, c.String())
		}
	}
}

func TestDecimal_Mul(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"2", "3"}:                     "6",
		Inp{"2454495034", "3451204593"}:   "8470964534836491162",
		Inp{"24544.95034", ".3451204593"}: "8470.964534836491162",
		Inp{".1", ".1"}:                   "0.01",
		Inp{"0", "1.001"}:                 "0",
	}

	for inp, res := range inputs {
		a, err := NewFromString(inp.a)
		if err != nil {
			t.FailNow()
		}
		b, err := NewFromString(inp.b)
		if err != nil {
			t.FailNow()
		}
		c := a.Mul(b)
		if c.String() != res {
			t.Errorf("expected %s, got %s", res, c.String())
		}
	}

	// positive scale
	c := New(1234, 5).Mul(New(45, -1))
	if c.String() != "555300000" {
		t.Errorf("Expected %s, got %s", "555300000", c.String())
	}
}

func TestDecimal_Div(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"6", "3"}:                            "2",
		Inp{"10", "2"}:                           "5",
		Inp{"2.2", "1.1"}:                        "2",
		Inp{"-2.2", "-1.1"}:                      "2",
		Inp{"12.88", "5.6"}:                      "2.3",
		Inp{"1023427554493", "43432632"}:         "23563.5628642767953828", // rounded
		Inp{"1", "434324545566634"}:              "0.0000000000000023",
		Inp{"1", "3"}:                            "0.3333333333333333",
		Inp{"2", "3"}:                            "0.6666666666666667", // rounded
		Inp{"10000", "3"}:                        "3333.3333333333333333",
		Inp{"10234274355545544493", "-3"}:        "-3411424785181848164.3333333333333333",
		Inp{"-4612301402398.4753343454", "23.5"}: "-196268144782.9138440146978723",
	}

	for inp, expected := range inputs {
		num, err := NewFromString(inp.a)
		if err != nil {
			t.FailNow()
		}
		denom, err := NewFromString(inp.b)
		if err != nil {
			t.FailNow()
		}
		got := num.Div(denom)
		if got.String() != expected {
			t.Errorf("expected %s when dividing %v by %v, got %v",
				expected, num, denom, got)
		}
	}

	type Inp2 struct {
		n    int64
		exp  int32
		n2   int64
		exp2 int32
	}

	// test code path where exp > 0
	inputs2 := map[Inp2]string{
		Inp2{124, 10, 3, 1}: "41333333333.3333333333333333",
		Inp2{124, 10, 3, 0}: "413333333333.3333333333333333",
		Inp2{124, 10, 6, 1}: "20666666666.6666666666666667",
		Inp2{124, 10, 6, 0}: "206666666666.6666666666666667",
		Inp2{10, 10, 10, 1}: "1000000000",
	}

	for inp, expectedAbs := range inputs2 {
		for i := -1; i <= 1; i += 2 {
			for j := -1; j <= 1; j += 2 {
				n := inp.n * int64(i)
				n2 := inp.n2 * int64(j)
				num := New(n, inp.exp)
				denom := New(n2, inp.exp2)
				expected := expectedAbs
				if i != j {
					expected = "-" + expectedAbs
				}
				got := num.Div(denom)
				if got.String() != expected {
					t.Errorf("expected %s when dividing %v by %v, got %v",
						expected, num, denom, got)
				}
			}
		}
	}
}

func TestDecimal_Mod(t *testing.T) {
	type Inp struct {
		a string
		b string
	}

	inputs := map[Inp]string{
		Inp{"3", "2"}:                     "1",
		Inp{"3451204593", "2454495034"}:   "996709559",
		Inp{"24544.95034", ".3451204593"}: "0.3283950433",
		Inp{".1", ".1"}:                   "0",
		Inp{"0", "1.001"}:                 "0",
		Inp{"-7.5", "2"}:                  "-1.5",
		Inp{"7.5", "-2"}:                  "1.5",
		Inp{"-7.5", "-2"}:                 "-1.5",
	}

	for inp, res := range inputs {
		a, err := NewFromString(inp.a)
		if err != nil {
			t.FailNow()
		}
		b, err := NewFromString(inp.b)
		if err != nil {
			t.FailNow()
		}
		c := a.Mod(b)
		if c.String() != res {
			t.Errorf("expected %s, got %s", res, c.String())
		}
	}
}

func TestDecimal_Overflow(t *testing.T) {
	if !didPanic(func() { New(1, math.MinInt32).Mul(New(1, math.MinInt32)) }) {
		t.Fatalf("should have gotten an overflow panic")
	}
	if !didPanic(func() { New(1, math.MaxInt32).Mul(New(1, math.MaxInt32)) }) {
		t.Fatalf("should have gotten an overflow panic")
	}
}

func TestDecimal_ExtremeValues(t *testing.T) {
	// NOTE(vadim): this test takes pretty much forever
	if testing.Short() {
		t.Skip()
	}

	// NOTE(vadim): Seriously, the numbers invovled are so large that this
	// test will take way too long, so mark it as success if it takes over
	// 1 second. The way this test typically fails (integer overflow) is that
	// a wrong result appears quickly, so if it takes a long time then it is
	// probably working properly.
	// Why even bother testing this? Completeness, I guess. -Vadim
	const timeLimit = 1 * time.Second
	test := func(f func()) {
		c := make(chan bool)
		go func() {
			f()
			close(c)
		}()
		select {
		case <-c:
		case <-time.After(timeLimit):
		}
	}

	test(func() {
		got := New(123, math.MinInt32).Floor()
		if !got.Equals(NewFromFloat(0)) {
			t.Errorf("Error: got %s, expected 0", got)
		}
	})
	test(func() {
		got := New(123, math.MinInt32).Ceil()
		if !got.Equals(NewFromFloat(1)) {
			t.Errorf("Error: got %s, expected 1", got)
		}
	})
	test(func() {
		got := New(123, math.MinInt32).Rat().FloatString(10)
		expected := "0.0000000000"
		if got != expected {
			t.Errorf("Error: got %s, expected %s", got, expected)
		}
	})
}

func TestIntPart(t *testing.T) {
	for _, testCase := range []struct {
		Dec     string
		IntPart int64
	}{
		{"0.01", 0},
		{"12.1", 12},
		{"9999.999", 9999},
		{"-32768.01234", -32768},
	} {
		d, err := NewFromString(testCase.Dec)
		if err != nil {
			t.Fatal(err)
		}
		if d.IntPart() != testCase.IntPart {
			t.Errorf("expect %d, got %d", testCase.IntPart, d.IntPart())
		}
	}
}

func TestDecimal_Min(t *testing.T) {
	// the first element in the array is the expected answer, rest are inputs
	testCases := [][]float64{
		{0, 0},
		{1, 1},
		{-1, -1},
		{1, 1, 2},
		{-2, 1, 2, -2},
		{-3, 0, 2, -2, -3},
	}

	for _, test := range testCases {
		expected, input := test[0], test[1:]
		expectedDecimal := NewFromFloat(expected)
		decimalInput := []Decimal{}
		for _, inp := range input {
			d := NewFromFloat(inp)
			decimalInput = append(decimalInput, d)
		}
		got := Min(decimalInput[0], decimalInput[1:]...)
		if !got.Equals(expectedDecimal) {
			t.Errorf("Expected %v, got %v, input=%+v", expectedDecimal, got,
				decimalInput)
		}
	}
}

func TestDecimal_Max(t *testing.T) {
	// the first element in the array is the expected answer, rest are inputs
	testCases := [][]float64{
		{0, 0},
		{1, 1},
		{-1, -1},
		{2, 1, 2},
		{2, 1, 2, -2},
		{3, 0, 3, -2},
		{-2, -3, -2},
	}

	for _, test := range testCases {
		expected, input := test[0], test[1:]
		expectedDecimal := NewFromFloat(expected)
		decimalInput := []Decimal{}
		for _, inp := range input {
			d := NewFromFloat(inp)
			decimalInput = append(decimalInput, d)
		}
		got := Max(decimalInput[0], decimalInput[1:]...)
		if !got.Equals(expectedDecimal) {
			t.Errorf("Expected %v, got %v, input=%+v", expectedDecimal, got,
				decimalInput)
		}
	}
}

func TestDecimal_Scan(t *testing.T) {
	// test the Scan method the implements the
	// sql.Scanner interface
	// check for the for different type of values
	// that are possible to be received from the database
	// drivers

	// in normal operations the db driver (sqlite at least)
	// will return an int64 if you specified a numeric format
	a := Decimal{}
	dbvalue := float64(54.33)
	expected := NewFromFloat(dbvalue)

	err := a.Scan(dbvalue)
	if err != nil {
		// Scan failed... no need to test result value
		t.Errorf("a.Scan(54.33) failed with message: %s", err)

	} else {
		// Scan suceeded... test resulting values
		if !a.Equals(expected) {
			t.Errorf("%s does not equal to %s", a, expected)
		}
	}

	// at least SQLite returns an int64 when 0 is stored in the db
	// and you specified a numeric format on the schema
	dbvalue_int := int64(0)
	expected = New(dbvalue_int, 0)

	err = a.Scan(dbvalue_int)
	if err != nil {
		// Scan failed... no need to test result value
		t.Errorf("a.Scan(0) failed with message: %s", err)

	} else {
		// Scan suceeded... test resulting values
		if !a.Equals(expected) {
			t.Errorf("%s does not equal to %s", a, expected)
		}
	}

	// in case you specified a varchar in your SQL schema,
	// the database driver will return byte slice []byte
	value_str := "535.666"
	dbvalue_str := []byte(value_str)
	expected, err = NewFromString(value_str)

	err = a.Scan(dbvalue_str)
	if err != nil {
		// Scan failed... no need to test result value
		t.Errorf("a.Scan('535.666') failed with message: %s", err)

	} else {
		// Scan suceeded... test resulting values
		if !a.Equals(expected) {
			t.Errorf("%s does not equal to %s", a, expected)
		}
	}
}

// old tests after this line

func TestDecimal_Scale(t *testing.T) {
	a := New(1234, -3)
	if a.Exponent() != -3 {
		t.Errorf("error")
	}
}

func TestDecimal_Abs1(t *testing.T) {
	a := New(-1234, -4)
	b := New(1234, -4)

	c := a.Abs()
	if c.Cmp(b) != 0 {
		t.Errorf("error")
	}
}

func TestDecimal_Abs2(t *testing.T) {
	a := New(-1234, -4)
	b := New(1234, -4)

	c := b.Abs()
	if c.Cmp(a) == 0 {
		t.Errorf("error")
	}
}

func TestDecimal_Equal(t *testing.T) {
	a := New(1234, 3)
	b := New(1234, 3)

	if !a.Equals(b) {
		t.Errorf("%q should equal %q", a, b)
	}
}

func TestDecimal_ScalesNotEqual(t *testing.T) {
	a := New(1234, 2)
	b := New(1234, 3)
	if a.Equals(b) {
		t.Errorf("%q should not equal %q", a, b)
	}
}

func TestDecimal_Cmp1(t *testing.T) {
	a := New(123, 3)
	b := New(-1234, 2)

	if a.Cmp(b) != 1 {
		t.Errorf("Error")
	}
}

func TestDecimal_Cmp2(t *testing.T) {
	a := New(123, 3)
	b := New(1234, 2)

	if a.Cmp(b) != -1 {
		t.Errorf("Error")
	}
}

func didPanic(f func()) bool {
	ret := false
	func() {

		defer func() {
			if message := recover(); message != nil {
				ret = true
			}
		}()

		// call the target function
		f()

	}()

	return ret

}

type DecimalSlice []Decimal

func (p DecimalSlice) Len() int           { return len(p) }
func (p DecimalSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p DecimalSlice) Less(i, j int) bool { return p[i].Cmp(p[j]) < 0 }
func Benchmark_Cmp(b *testing.B) {
	decimals := DecimalSlice([]Decimal{})
	for i := 0; i < 1000000; i++ {
		decimals = append(decimals, New(int64(i), 0))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sort.Sort(decimals)
	}
}
