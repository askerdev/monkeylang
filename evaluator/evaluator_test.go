package evaluator

import (
	"testing"

	"github.com/askerdev/monkeylang/lexer"
	"github.com/askerdev/monkeylang/object"
	"github.com/askerdev/monkeylang/parser"
)

func TestEvalStringExpression(t *testing.T) {
	tc := []struct {
		input    string
		expected string
	}{
		{`"hello" + " " + "world"`, "hello world"},
	}

	for _, tt := range tc {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
	}
}

func TestEvalIntegerExpression(t *testing.T) {
	tc := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tc {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tc := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{`"hello" == "hello"`, true},
	}

	for _, tt := range tc {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestBangOperator(t *testing.T) {
	tc := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
		{`!!"hello"`, true},
		{`!"world"`, false},
		{`!""`, false},
		{`!!""`, true},
	}

	for _, tt := range tc {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpression(t *testing.T) {
	tc := []struct {
		input    string
		expected any
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tt := range tc {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tc := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{`if (10 > 1) {
			if (10 > 1) {
				return 10;
			}
			return 1;
		}`, 10},
	}

	for _, tt := range tc {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tc := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
				if (10 > 1) {
					if (10 > 1) {
						return true + false;
					}
					return 1;
				}
			`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar;",
			"identifier not found: foobar",
		},
	}

	for i, tt := range tc {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("[%d] expected *object.Error, got %T(%+v)", i, evaluated, evaluated)
			continue
		}

		if errObj.Message != tt.expectedMessage {
			t.Errorf("[%d] expected error message %q, got %q", i, tt.expectedMessage, errObj.Message)
		}
	}
}

func TestLetStatements(t *testing.T) {
	tc := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tt := range tc {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := `fn(x) { x + 2; }`

	evaluated := testEval(input)

	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("expected *object.Function, got %T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("wrong number of parameters, expected 1 got %d", len(fn.Parameters))
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not x, got %q", fn.Parameters[0].String())
	}

	expectedBody := `(x + 2)`

	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q, got %q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tc := []struct {
		input    string
		expected int64
	}{
		{"let identify = fn(x) { x; }; identify(5);", 5},
		{"let identify = fn(x) { return x; }; identify(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(a, b) { a + b; }; add(5, 5);", 10},
		{"let add = fn(a, b) { a + b; }; add(5, add(5, 5));", 15},
		{"fn(x) { x; }(5)", 5},
	}

	for _, tt := range tc {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestClosures(t *testing.T) {
	input := `
	let newAdder = fn(x) {
	fn(y) { x + y };
	};
	let addTwo = newAdder(2);
	addTwo(2);`
	testIntegerObject(t, testEval(input), 4)
}

func TestBuiltinFunctions(t *testing.T) {
	tc := []struct {
		input    string
		expected any
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not supported, got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
	}

	for _, tt := range tc {
		evaluated := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)",
					evaluated, evaluated)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q",
					expected, errObj.Message)
			}
		}
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()
	return Eval(program, env)
}

func testStringObject(t *testing.T, obj object.Object, expected string) bool {
	result, ok := obj.(*object.String)
	if !ok {
		t.Errorf("obj is not *object.String, got %T", obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%q, want=%q",
			result.Value, expected)
		return false
	}

	return true
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("obj is not *object.Integer, got %T", obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d",
			result.Value, expected)
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t",
			result.Value, expected)
		return false
	}
	return true
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}
