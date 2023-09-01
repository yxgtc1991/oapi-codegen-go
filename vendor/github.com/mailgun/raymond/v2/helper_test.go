package raymond

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	VERBOSE = false
)

//
// Helpers
//

func barHelper(options *Options) string { return "bar" }

func echoHelper(str string, nb int) string {
	result := ""
	for i := 0; i < nb; i++ {
		result += str
	}

	return result
}

func boolHelper(b bool) string {
	if b {
		return "yes it is"
	}

	return "absolutely not"
}

func gnakHelper(nb int) string {
	result := ""
	for i := 0; i < nb; i++ {
		result += "GnAK!"
	}

	return result
}

func variadicHelper(values ...string) string {
	return strings.Join(values, ",")
}

//
// Tests
//

var helperTests = []Test{
	{
		"simple helper",
		`{{foo}}`,
		nil, nil,
		map[string]interface{}{"foo": barHelper},
		nil,
		`bar`,
	},
	{
		"helper with literal string param",
		`{{echo "foo" 1}}`,
		nil, nil,
		map[string]interface{}{"echo": echoHelper},
		nil,
		`foo`,
	},
	{
		"helper with identifier param",
		`{{echo foo 1}}`,
		map[string]interface{}{"foo": "bar"},
		nil,
		map[string]interface{}{"echo": echoHelper},
		nil,
		`bar`,
	},
	{
		"helper with literal boolean param",
		`{{bool true}}`,
		nil, nil,
		map[string]interface{}{"bool": boolHelper},
		nil,
		`yes it is`,
	},
	{
		"helper with literal boolean param",
		`{{bool false}}`,
		nil, nil,
		map[string]interface{}{"bool": boolHelper},
		nil,
		`absolutely not`,
	},
	{
		"helper with literal boolean param",
		`{{gnak 5}}`,
		nil, nil,
		map[string]interface{}{"gnak": gnakHelper},
		nil,
		`GnAK!GnAK!GnAK!GnAK!GnAK!`,
	},
	{
		"helper with several parameters",
		`{{echo "GnAK!" 3}}`,
		nil, nil,
		map[string]interface{}{"echo": echoHelper},
		nil,
		`GnAK!GnAK!GnAK!`,
	},
	{
		"helper with variadic parameters",
		`{{join "foo" "bar" "baz"}}`,
		nil, nil,
		map[string]interface{}{"join": variadicHelper},
		nil,
		`foo,bar,baz`,
	},
	{
		"#if helper with true literal",
		`{{#if true}}YES MAN{{/if}}`,
		nil, nil, nil, nil,
		`YES MAN`,
	},
	{
		"#if helper with false literal",
		`{{#if false}}YES MAN{{/if}}`,
		nil, nil, nil, nil,
		``,
	},
	{
		"#if helper with truthy identifier",
		`{{#if ok}}YES MAN{{/if}}`,
		map[string]interface{}{"ok": true},
		nil, nil, nil,
		`YES MAN`,
	},
	{
		"#if helper with falsy identifier",
		`{{#if ok}}YES MAN{{/if}}`,
		map[string]interface{}{"ok": false},
		nil, nil, nil,
		``,
	},
	{
		"#unless helper with true literal",
		`{{#unless true}}YES MAN{{/unless}}`,
		nil, nil, nil, nil,
		``,
	},
	{
		"#unless helper with false literal",
		`{{#unless false}}YES MAN{{/unless}}`,
		nil, nil, nil, nil,
		`YES MAN`,
	},
	{
		"#unless helper with truthy identifier",
		`{{#unless ok}}YES MAN{{/unless}}`,
		map[string]interface{}{"ok": true},
		nil, nil, nil,
		``,
	},
	{
		"#unless helper with falsy identifier",
		`{{#unless ok}}YES MAN{{/unless}}`,
		map[string]interface{}{"ok": false},
		nil, nil, nil,
		`YES MAN`,
	},
	{
		"#equal helper with same string var",
		`{{#equal foo "bar"}}YES MAN{{/equal}}`,
		map[string]interface{}{"foo": "bar"},
		nil, nil, nil,
		`YES MAN`,
	},
	{
		"#equal helper with different string var",
		`{{#equal foo "baz"}}YES MAN{{/equal}}`,
		map[string]interface{}{"foo": "bar"},
		nil, nil, nil,
		``,
	},
	{
		"#equal helper with same string vars",
		`{{#equal foo bar}}YES MAN{{/equal}}`,
		map[string]interface{}{"foo": "baz", "bar": "baz"},
		nil, nil, nil,
		`YES MAN`,
	},
	{
		"#equal helper with different string vars",
		`{{#equal foo bar}}YES MAN{{/equal}}`,
		map[string]interface{}{"foo": "baz", "bar": "tag"},
		nil, nil, nil,
		``,
	},
	{
		"#equal helper with same integer var",
		`{{#equal foo 1}}YES MAN{{/equal}}`,
		map[string]interface{}{"foo": 1},
		nil, nil, nil,
		`YES MAN`,
	},
	{
		"#equal helper with different integer var",
		`{{#equal foo 0}}YES MAN{{/equal}}`,
		map[string]interface{}{"foo": 1},
		nil, nil, nil,
		``,
	},
	{
		"#ifGt helper with true literal",
		`{{#ifGt foo 10}}foo is greater than 10{{/ifGt}}`,
		map[string]interface{}{"foo": 11},
		nil, nil, nil,
		`foo is greater than 10`,
	},
	{
		"#ifGt helper with false literal",
		`{{#ifGt foo 10}}foo is greater than 10{{/ifGt}}`,
		map[string]interface{}{"foo": 5},
		nil, nil, nil,
		``,
	},
	{
		"#ifGt helper with true literal from params",
		`{{#ifGt foo bar}}foo is greater than bar{{/ifGt}}`,
		map[string]interface{}{"foo": 5, "bar": 0},
		nil, nil, nil,
		`foo is greater than bar`,
	},
	{
		"#ifGt helper with string comparison",
		`{{#ifGt foo bar}}foo is greater than bar{{/ifGt}}`,
		map[string]interface{}{"foo": "5", "bar": "0"},
		nil, nil, nil,
		`foo is greater than bar`,
	},
	{
		"#ifGt helper with false literal from params",
		`{{#ifGt foo bar}}foo is greater than bar{{/ifGt}}`,
		map[string]interface{}{"foo": 5, "bar": 0},
		nil, nil, nil,
		`foo is greater than bar`,
	},
	{
		"#ifGt helper with else condition",
		`{{#ifGt foo bar}}foo is greater than bar{{else}}foo is not greater than bar{{/ifGt}}`,
		map[string]interface{}{"foo": 0, "bar": 5},
		nil, nil, nil,
		`foo is not greater than bar`,
	},
	{
		"#ifGt helper non-numbers",
		`{{#ifGt foo bar}}foo is greater than bar{{/ifGt}}`,
		map[string]interface{}{"foo": "foo", "bar": "bar"},
		nil, nil, nil,
		``,
	},
	{
		"#ifGt helper non-numbers with else condition",
		`{{#ifGt foo bar}}foo is greater than bar{{else}}foo or bar are not numbers{{/ifGt}}`,
		map[string]interface{}{"foo": "foo", "bar": "bar"},
		nil, nil, nil,
		`foo or bar are not numbers`,
	},
	{
		"#ifLt helper with true literal",
		`{{#ifLt foo 10}}foo is less than 10{{/ifLt}}`,
		map[string]interface{}{"foo": 5},
		nil, nil, nil,
		`foo is less than 10`,
	},
	{
		"#ifLt helper with false literal",
		`{{#ifLt foo 10}}foo is less than 10{{/ifLt}}`,
		map[string]interface{}{"foo": 15},
		nil, nil, nil,
		``,
	},
	{
		"#ifLt helper with true literal from params",
		`{{#ifLt foo bar}}foo is less than bar{{/ifLt}}`,
		map[string]interface{}{"foo": 0, "bar": 5},
		nil, nil, nil,
		`foo is less than bar`,
	},
	{
		"#ifLt helper with string comparison",
		`{{#ifLt foo bar}}foo is less than bar{{/ifLt}}`,
		map[string]interface{}{"foo": "0", "bar": "5"},
		nil, nil, nil,
		`foo is less than bar`,
	},
	{
		"#ifLt helper with false literal from params",
		`{{#ifLt foo bar}}foo is less than bar{{/ifLt}}`,
		map[string]interface{}{"foo": 0, "bar": 5},
		nil, nil, nil,
		`foo is less than bar`,
	},
	{
		"#ifLt helper with else condition",
		`{{#ifLt foo bar}}foo is less than bar{{else}}foo is not less than bar{{/ifLt}}`,
		map[string]interface{}{"foo": 6, "bar": 5},
		nil, nil, nil,
		`foo is not less than bar`,
	},
	{
		"#ifLt helper non-numbers",
		`{{#ifLt foo bar}}foo is less than bar{{/ifLt}}`,
		map[string]interface{}{"foo": "foo", "bar": "bar"},
		nil, nil, nil,
		``,
	},
	{
		"#ifLt helper non-numbers with else condition",
		`{{#ifLt foo bar}}foo is greater than bar{{else}}foo or bar are not numbers{{/ifLt}}`,
		map[string]interface{}{"foo": "foo", "bar": "bar"},
		nil, nil, nil,
		`foo or bar are not numbers`,
	},
	{
		"#ifEq helper with true literal",
		`{{#ifEq foo 10}}foo is equal to 10{{/ifEq}}`,
		map[string]interface{}{"foo": 10},
		nil, nil, nil,
		`foo is equal to 10`,
	},
	{
		"#ifEq helper with false literal",
		`{{#ifEq foo 10}}foo is equal to 10{{/ifEq}}`,
		map[string]interface{}{"foo": 15},
		nil, nil, nil,
		``,
	},
	{
		"#ifEq helper with true literal from params",
		`{{#ifEq foo bar}}foo is equal to bar{{/ifEq}}`,
		map[string]interface{}{"foo": 5, "bar": 5},
		nil, nil, nil,
		`foo is equal to bar`,
	},
	{
		"#ifEq helper with string comparison",
		`{{#ifEq foo bar}}foo is equal to bar{{/ifEq}}`,
		map[string]interface{}{"foo": "5", "bar": "5"},
		nil, nil, nil,
		`foo is equal to bar`,
	},
	{
		"#ifEq helper with false literal from params",
		`{{#ifEq foo bar}}foo is equal to bar{{/ifEq}}`,
		map[string]interface{}{"foo": 5, "bar": 5},
		nil, nil, nil,
		`foo is equal to bar`,
	},
	{
		"#ifEq helper with else condition",
		`{{#ifEq foo bar}}foo is equal to bar{{else}}foo is not equal to bar{{/ifEq}}`,
		map[string]interface{}{"foo": 6, "bar": 5},
		nil, nil, nil,
		`foo is not equal to bar`,
	},
	{
		"#ifEq helper non-numbers",
		`{{#ifEq foo bar}}foo is equal to bar{{/ifEq}}`,
		map[string]interface{}{"foo": "foo", "bar": "bar"},
		nil, nil, nil,
		``,
	},
	{
		"#ifEq helper non-numbers with else condition",
		`{{#ifEq foo bar}}foo is equal to bar{{else}}foo or bar are not numbers{{/ifEq}}`,
		map[string]interface{}{"foo": "foo", "bar": "bar"},
		nil, nil, nil,
		`foo or bar are not numbers`,
	},
	{
		"#ifMatchesRegexStr helper simple string match",
		`{{#ifMatchesRegexStr "foo" bar}}The expression matches{{/ifMatchesRegexStr}}`,
		map[string]interface{}{"bar": "foo"},
		nil, nil, nil,
		`The expression matches`,
	},
	{
		"#ifMatchesRegexStr regex match",
		`{{#ifMatchesRegexStr regex phone}}The expression matches{{/ifMatchesRegexStr}}`,
		map[string]interface{}{
			"regex": "^(\\+\\d{1,2}\\s)?\\(?\\d{3}\\)?[\\s.-]\\d{3}[\\s.-]\\d{4}$",
			"phone": "555-333-4545",
		},
		nil, nil, nil,
		`The expression matches`,
	},
	{
		"#ifMatchesRegexStr regex match with string conversion",
		`{{#ifMatchesRegexStr regex phone}}The expression matches{{/ifMatchesRegexStr}}`,
		map[string]interface{}{
			"regex": "^(\\+\\d{1,2}\\s)?\\(?\\d{3}\\)?\\d{3}\\d{4}$",
			"phone": 5553334545,
		},
		nil, nil, nil,
		`The expression matches`,
	},
	{
		"#ifMatchesRegexStr helper simple string does not match",
		`{{#ifMatchesRegexStr "foo" bar}}The expression matches{{/ifMatchesRegexStr}}`,
		map[string]interface{}{"bar": "bar"},
		nil, nil, nil,
		``,
	},
	{
		"#ifMatchesRegexStr regex does not match",
		`{{#ifMatchesRegexStr regex phone}}The expression matches{{/ifMatchesRegexStr}}`,
		map[string]interface{}{
			"regex": "^(\\+\\d{1,2}\\s)?\\(?\\d{3}\\)?[\\s.-]\\d{3}[\\s.-]\\d{4}$",
			"phone": "5553334545",
		},
		nil, nil, nil,
		``,
	},
	{
		"#ifMatchesRegexStr regex does not match with string conversion",
		`{{#ifMatchesRegexStr regex phone}}The expression matches{{/ifMatchesRegexStr}}`,
		map[string]interface{}{
			"regex": "^(\\+\\d{1,2}\\s)?\\(?\\d{3}\\)?\\d{3}\\d{4}$",
			"phone": 1,
		},
		nil, nil, nil,
		``,
	},
	{
		"#ifMatchesRegexStr helper simple string trigger else",
		`{{#ifMatchesRegexStr "foo" bar}}The expression matches{{else}}The expression does not match{{/ifMatchesRegexStr}}`,
		map[string]interface{}{"bar": "bar"},
		nil, nil, nil,
		`The expression does not match`,
	},
	{
		"#ifMatchesRegexStr regex trigger else",
		`{{#ifMatchesRegexStr regex phone}}The expression matches{{else}}The expression does not match{{/ifMatchesRegexStr}}`,
		map[string]interface{}{
			"regex": "^(\\+\\d{1,2}\\s)?\\(?\\d{3}\\)?[\\s.-]\\d{3}[\\s.-]\\d{4}$",
			"phone": "5553334545",
		},
		nil, nil, nil,
		`The expression does not match`,
	},
	{
		"#ifMatchesRegexStr regex trigger else with string conversion",
		`{{#ifMatchesRegexStr regex phone}}The expression matches{{else}}The expression does not match{{/ifMatchesRegexStr}}`,
		map[string]interface{}{
			"regex": "^(\\+\\d{1,2}\\s)?\\(?\\d{3}\\)?\\d{3}\\d{4}$",
			"phone": 1,
		},
		nil, nil, nil,
		`The expression does not match`,
	},
	{
		"length helper on map",
		"{{#ifEq foo.length 3}}Length is equal to 3{{/ifEq}}",
		map[string]interface{}{"foo": map[string]string{
			"rick":   "bird person",
			"beth":   "jerry",
			"summer": "morty",
		},
		},
		nil, nil, nil,
		`Length is equal to 3`,
	},
	{
		"length helper on slice",
		"{{#ifEq foo.length 2}}Length is equal to 2{{/ifEq}}",
		map[string]interface{}{"foo": []string{"foo", "bar"}},
		nil, nil, nil,
		`Length is equal to 2`,
	},
	{
		"length helper on string",
		"{{#ifEq foo.length 3}}Length is equal to 3{{/ifEq}}",
		map[string]interface{}{"foo": "bar"},
		nil, nil, nil,
		`Length is equal to 3`,
	},
	{
		"length helper on map false condition",
		"{{#ifEq foo.length 4}}Length is equal to 4{{/ifEq}}",
		map[string]interface{}{"foo": map[string]string{
			"rick":   "bird person",
			"beth":   "jerry",
			"summer": "morty",
		},
		},
		nil, nil, nil,
		``,
	},
	{
		"length helper on slice false condition",
		"{{#ifEq foo.length 4}}Length is equal to 4{{/ifEq}}",
		map[string]interface{}{"foo": []string{"foo", "bar"}},
		nil, nil, nil,
		``,
	},
	{
		"length helper on string false condition",
		"{{#ifEq foo.length 4}}Length is equal to 4{{/ifEq}}",
		map[string]interface{}{"foo": "bar"},
		nil, nil, nil,
		``,
	},
	{
		"length helper on map else condition",
		"{{#ifEq foo.length 4}}Length is equal to 4{{else}}Length is not equal to 4{{/ifEq}}",
		map[string]interface{}{"foo": map[string]string{
			"rick":   "bird person",
			"beth":   "jerry",
			"summer": "morty",
		},
		},
		nil, nil, nil,
		`Length is not equal to 4`,
	},
	{
		"length helper on slice else condition",
		"{{#ifEq foo.length 4}}Length is equal to 4{{else}}Length is not equal to 4{{/ifEq}}",
		map[string]interface{}{"foo": []string{"foo", "bar"}},
		nil, nil, nil,
		`Length is not equal to 4`,
	},
	{
		"length helper on string else condition",
		"{{#ifEq foo.length 4}}Length is equal to 4{{else}}Length is not equal to 4{{/ifEq}}",
		map[string]interface{}{"foo": "bar"},
		nil, nil, nil,
		`Length is not equal to 4`,
	},
	{
		"pluralize result plural",
		`{{pluralize error_count "errors" "error"}}`,
		map[string]interface{}{"error_count": 3},
		nil, nil, nil,
		`errors`,
	},
	{
		"pluralize result singular",
		`{{pluralize error_count "errors" "error"}}`,
		map[string]interface{}{"error_count": 1},
		nil, nil, nil,
		`error`,
	},
	{
		"pluralize with vars result plural",
		`{{pluralize error.count error.plural error.singular}}`,
		map[string]interface{}{"error": map[string]interface{}{
			"count":    3,
			"plural":   "errors",
			"singular": "error",
		}},
		nil, nil, nil,
		`errors`,
	},
	{
		"pluralize with vars result singular",
		`{{pluralize error.count error.plural error.singular}}`,
		map[string]interface{}{"error": map[string]interface{}{
			"count":    1,
			"plural":   "errors",
			"singular": "error",
		}},
		nil, nil, nil,
		`error`,
	},
	{
		"pluralize count non-number",
		`{{pluralize error_count "errors" "error"}}`,
		map[string]interface{}{"error_count": "three"},
		nil, nil, nil,
		`error`,
	},
	{
		"#equal helper inside HTML tag",
		`<option value="test" {{#equal value "test"}}selected{{/equal}}>Test</option>`,
		map[string]interface{}{"value": "test"},
		nil, nil, nil,
		`<option value="test" selected>Test</option>`,
	},
	{
		"#equal full example",
		`{{#equal foo "bar"}}foo is bar{{/equal}}
{{#equal foo baz}}foo is the same as baz{{/equal}}
{{#equal nb 0}}nothing{{/equal}}
{{#equal nb 1}}there is one{{/equal}}
{{#equal nb "1"}}everything is stringified before comparison{{/equal}}`,
		map[string]interface{}{
			"foo": "bar",
			"baz": "bar",
			"nb":  1,
		},
		nil, nil, nil,
		`foo is bar
foo is the same as baz

there is one
everything is stringified before comparison`,
	},
}

//
// Let's go
//

func TestHelper(t *testing.T) {
	t.Parallel()

	launchTests(t, helperTests)
}

func TestRemoveHelper(t *testing.T) {
	RegisterHelper("testremovehelper", func() string { return "" })
	if _, ok := helpers["testremovehelper"]; !ok {
		t.Error("Failed to register global helper")
	}

	RemoveHelper("testremovehelper")
	if _, ok := helpers["testremovehelper"]; ok {
		t.Error("Failed to remove global helper")
	}
}

//
// Fixes: https://github.com/aymerick/raymond/issues/2
//

type Author struct {
	FirstName string
	LastName  string
}

func TestHelperCtx(t *testing.T) {
	RegisterHelper("template", func(name string, options *Options) SafeString {
		context := options.Ctx()

		template := name + " - {{ firstName }} {{ lastName }}"
		result, _ := Render(template, context)

		return SafeString(result)
	})

	template := `By {{ template "namefile" }}`
	context := Author{"Alan", "Johnson"}

	result, _ := Render(template, context)
	if result != "By namefile - Alan Johnson" {
		t.Errorf("Failed to render template in helper: %q", result)
	}
}

func TestRegisterParamHelper(t *testing.T) {
	pHelper := func(v reflect.Value) reflect.Value { return v }
	RegisterParamHelper("test", pHelper)

	gotFunc := findParamHelper("test")
	require.NotNil(t, gotFunc)

	value := reflect.ValueOf("rick")
	got := gotFunc(value)
	assert.Equal(t, value.String(), got.String())

	RemoveParamHelper("test")
	gotFunc = findParamHelper("test")
	assert.Nil(t, gotFunc)
}
