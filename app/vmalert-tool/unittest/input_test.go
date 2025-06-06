package unittest

import (
	"math"
	"testing"
	"time"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/decimal"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/promutil"
)

func TestParseInputValue_Failure(t *testing.T) {
	f := func(input string) {
		t.Helper()

		val, err := parseInputValue(input, true)
		if err == nil {
			t.Fatalf("expecting non-nil error, got value %+v", val)
		}
	}

	f("")
	f("x4")
	f("testfailed")

	// stale doesn't support operations
	f("stalex3")

	f("1+stalex3")
}

func TestParseInputValue_Success(t *testing.T) {
	f := func(input string, outputExpected []sequenceValue) {
		t.Helper()

		output, err := parseInputValue(input, true)
		if err != nil {
			t.Fatalf("unexpected error in parseInputValue: %s", err)
		}

		if len(outputExpected) != len(output) {
			t.Fatalf("unexpected output length; got %d; want %d", len(outputExpected), len(output))
		}
		for i := 0; i < len(outputExpected); i++ {
			if outputExpected[i].Omitted != output[i].Omitted {
				t.Fatalf("unexpected Omitted field in the output\ngot\n%v\nwant\n%v", output, outputExpected)
			}
			if outputExpected[i].Value != output[i].Value {
				if (math.IsNaN(outputExpected[i].Value) && math.IsNaN(output[i].Value)) ||
					(decimal.IsStaleNaN(outputExpected[i].Value) && decimal.IsStaleNaN(output[i].Value)) {
					continue
				}
				t.Fatalf("unexpected Value field in the output\ngot\n%v\nwant\n%v", output, outputExpected)
			}
		}
	}

	f("-4", []sequenceValue{{Value: -4}})

	f("_", []sequenceValue{{Omitted: true}})

	f("stale", []sequenceValue{{Value: decimal.StaleNaN}})

	f(" stale    2", []sequenceValue{{Value: decimal.StaleNaN}, {Value: 2}})

	f("-4x1", []sequenceValue{{Value: -4}, {Value: -4}})

	f("_x1", []sequenceValue{{Omitted: true}})

	f("1+1x2 0.1 0.1+0.3x2 3.14", []sequenceValue{{Value: 1}, {Value: 2}, {Value: 3}, {Value: 0.1}, {Value: 0.1}, {Value: 0.4}, {Value: 0.7}, {Value: 3.14}})

	f("2-1x4", []sequenceValue{{Value: 2}, {Value: 1}, {Value: 0}, {Value: -1}, {Value: -2}})

	f("1+1x1 _ -4 stale 3+20x1", []sequenceValue{{Value: 1}, {Value: 2}, {Omitted: true}, {Value: -4}, {Value: decimal.StaleNaN}, {Value: 3}, {Value: 23}})

	f("Inf +Inf -Inf", []sequenceValue{{Value: math.Inf(1)}, {Value: math.Inf(1)}, {Value: math.Inf(-1)}})

	f("Nan Infx2", []sequenceValue{{Value: math.NaN()}, {Value: math.Inf(1)}, {Value: math.Inf(1)}, {Value: math.Inf(1)}})

	f("NaNx2", []sequenceValue{{Value: math.NaN()}, {Value: math.NaN()}, {Value: math.NaN()}})

	f("4e9+1x2 -3e-2x2 1e-2-1x1", []sequenceValue{{Value: 4000000000}, {Value: 4000000001}, {Value: 4000000002}, {Value: -0.03}, {Value: -0.03}, {Value: -0.03}, {Value: 0.01}, {Value: -0.99}})
}

func TestParseInputSeries_Success(t *testing.T) {
	f := func(input []series) {
		t.Helper()
		var interval promutil.Duration
		_, err := parseInputSeries(input, &interval, time.Now())
		if err != nil {
			t.Fatalf("expect to see no error: %v", err)
		}
	}

	f([]series{{Series: "test", Values: "1"}})
	f([]series{{Series: "test{}", Values: "1"}})
	f([]series{{Series: "test{env=\"prod\",job=\"a\" }", Values: "1"}})
	f([]series{{Series: "{__name__=\"test\",env=\"prod\",job=\"a\" }", Values: "1"}})
}

func TestParseInputSeries_Fail(t *testing.T) {
	f := func(input []series) {
		t.Helper()
		var interval promutil.Duration
		_, err := parseInputSeries(input, &interval, time.Now())
		if err == nil {
			t.Fatalf("expect to see error: %v", err)
		}
	}

	f([]series{{Series: "", Values: "1"}})
	f([]series{{Series: "{}", Values: "1"}})
	f([]series{{Series: "{env=\"prod\",job=\"a\" or env=\"dev\",job=\"b\"}", Values: "1"}})
}
