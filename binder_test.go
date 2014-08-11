package bind_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/robfig/bind"
)

type bindTest struct {
	name     string
	params   map[string][]string
	expected interface{}
}
type p map[string][]string

var (
	testDate     = time.Date(1982, time.July, 9, 0, 0, 0, 0, time.UTC)
	testDatetime = time.Date(1982, time.July, 9, 21, 30, 0, 0, time.UTC)
)

func init() {
	bind.TimeFormats = append(bind.TimeFormats, "01/02/2006")
}

func pint(val int) *int {
	return &val
}

var mk = func(name, val string, expected interface{}) bindTest {
	return bindTest{name, map[string][]string{name: []string{val}}, expected}
}

var bindValueTests = []bindTest{
	mk("int", "1", int(1)),
	mk("int", "0", int(0)),
	mk("int", "-1", int(-1)),
	mk("int8", "1", int8(1)),
	mk("int16", "1", int16(1)),
	mk("int32", "1", int32(1)),
	mk("int64", "1", int64(1)),
	mk("uint", "1", uint(1)),
	mk("uint8", "1", uint8(1)),
	mk("uint16", "1", uint16(1)),
	mk("uint32", "1", uint32(1)),
	mk("uint64", "1", uint64(1)),
	mk("float32", "1.000000", float32(1.0)),
	mk("float64", "1.000000", float64(1.0)),
	mk("str", "hello", "hello"),
	mk("bool-true", "true", true),
	mk("bool-1", "1", true),
	mk("bool-on", "on", true),
	mk("bool-false", "false", false),
	mk("bool-0", "0", false),
	mk("bool-off", "", false),
	mk("date", "1982-07-09", testDate),
	mk("datetime", "1982-07-09 21:30", testDatetime),
	mk("customDate", "07/09/1982", testDate),
	mk("ptrint", "1", pint(1)),

	mk("errInvalidInt", "xyz", 0),
	mk("errInvalidInt2", "", 0),
	mk("errInvalidBool", "xyz", false),
	mk("errInt8-overflow", "1024", int8(0)),
	mk("errUint8-overflow", "1024", int8(0)),
	mk("errUint8-underflow", "-1", uint8(0)),
}

func TestBindValues(t *testing.T) {
	runBindTests(t, bindValueTests)
}

var bindSliceTests = []bindTest{
	{"arr", p{
		"arr[0]": {"1"},
		"arr[1]": {"2"},
		"arr[3]": {"3"},
	}, []int{1, 2, 0, 3}},

	{"arr", p{
		"arr": {"1", "2", "3"},
	}, []int{1, 2, 3}},

	{"uarr", p{
		"uarr[]": {"1", "2"},
	}, []uint{1, 2}},

	{"arruarr", p{
		"arruarr[0][]": {"1", "2"},
		"arruarr[1][]": {"3", "4"},
	}, [][]int{{1, 2}, {3, 4}}},

	{"2darr", p{
		"2darr[0][0]": {"0"},
		"2darr[0][1]": {"1"},
		"2darr[1][0]": {"10"},
		"2darr[1][1]": {"11"},
	}, [][]int{{0, 1}, {10, 11}}},
}

func TestBindSlices(t *testing.T) {
	runBindTests(t, bindSliceTests)
}

type B struct {
	Extra string
}
type A struct {
	Id      int
	Name    string
	B       B
	PB      *B
	private int
}

var bindStructTests = []bindTest{
	{"A", p{
		"A.Id":   {"123"},
		"A.Name": {"rob"},
	}, A{Id: 123, Name: "rob"}},

	{"B", p{
		"B.Id":      {"123"},
		"B.Name":    {"rob"},
		"B.B.Extra": {"hello"},
	}, A{Id: 123, Name: "rob", B: B{Extra: "hello"}}},

	{"pAB", p{
		"pAB.Id":      {"123"},
		"pAB.Name":    {"rob"},
		"pAB.B.Extra": {"hello"},
	}, &A{Id: 123, Name: "rob", B: B{Extra: "hello"}}},

	{"pB", p{
		"pB.Id":       {"123"},
		"pB.Name":     {"rob"},
		"pB.PB.Extra": {"hello"},
	}, &A{Id: 123, Name: "rob", PB: &B{Extra: "hello"}}},

	{"arrC", p{
		"arrC[0].Id":      {"5"},
		"arrC[0].Name":    {"rob"},
		"arrC[0].B.Extra": {"foo"},
		"arrC[1].Id":      {"8"},
		"arrC[1].Name":    {"bill"},
	}, []A{
		{
			Id:   5,
			Name: "rob",
			B:    B{"foo"},
		},
		{
			Id:   8,
			Name: "bill",
		},
	}},

	{"errPrivate", p{
		"errPriv.private": {"123"},
	}, A{}},
}

func TestBindStructs(t *testing.T) {
	runBindTests(t, bindStructTests)
}

func TestBindAll(t *testing.T) {
	var obj struct {
		Id     int32
		Labels []string
		Pets   []struct {
			Name string
		}
	}
	var err = bind.Values(map[string][]string{
		"Id":           {"5"},
		"Labels":       {"foo", "bar"},
		"Pets[0].Name": {"Lassie"},
		"Pets[1].Name": {"Mabel"},
	}).All(&obj)

	if err != nil {
		t.Error(err)
	}
	if obj.Id != 5 ||
		!reflect.DeepEqual(obj.Labels, []string{"foo", "bar"}) ||
		obj.Pets[0].Name != "Lassie" ||
		obj.Pets[1].Name != "Mabel" {
		t.Errorf("Wrong data, got %#v", obj)
	}
}

// TestNilPointerDestination verifies that it returns an appropriate error if a
// nil non-addressable pointer is passed, or that we new up a destination if we
// get the address of a nil pointer.
func TestNilPointerDestination(t *testing.T) {
	var params = map[string]string{
		"int": "5",
	}
	var actual *int
	var err = bind.Map(params).Field(actual, "int")
	if err == nil {
		t.Error("Expected an error: can not set the destination")
	}

	err = bind.Map(params).Field(&actual, "int")
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkBinder(b *testing.B) {
	var tests []bindTest
	tests = append(tests, bindValueTests...)
	tests = append(tests, bindSliceTests...)
	tests = append(tests, bindStructTests...)
	for i := 0; i < b.N; i++ {
		runBindTests(b, tests)
	}
}

func runBindTests(t testing.TB, tests []bindTest) {
	for _, test := range tests {
		var binder = bind.Values(test.params)
		var pactual = reflect.New(reflect.TypeOf(test.expected))
		var err = binder.Field(pactual.Interface(), test.name)
		if err != nil {
			if !strings.HasPrefix(test.name, "err") {
				t.Errorf("%v: %v", test.name, err)
			}
			continue
		}

		var actual = pactual.Elem().Interface()
		if !reflect.DeepEqual(actual, test.expected) {
			t.Errorf("%v: expected %#v, got %#v", test.name, test.expected, actual)
		}
	}
}

func recordErrors(t *testing.T, errs ...error) {
	for _, err := range errs {
		if err != nil {
			t.Error(err)
		}
	}
}

func TestMultipartFiles(t *testing.T) {
	const multipartFormData = `--A
Content-Disposition: form-data; name="text1"

data1
--A
Content-Disposition: form-data; name="file1"; filename="test.txt"
Content-Type: text/plain

content1
--A
Content-Disposition: form-data; name="file2[]"; filename="test.txt"
Content-Type: text/plain

content2
--A
Content-Disposition: form-data; name="file2[]"; filename="favicon.ico"
Content-Type: image/x-icon

xyz
--A--
`

	var req, _ = http.NewRequest("POST", "http://localhost/path",
		bytes.NewBufferString(multipartFormData))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=A")
	req.Header.Set("Content-Length", strconv.Itoa(len(multipartFormData)))
	var binder = bind.Request(req)

	// Test simple values
	var text1 string
	recordErrors(t, binder.Field(&text1, "text1"))
	if text1 != "data1" {
		t.Error("text1: got %v", text1)
	}

	// Test the files
	// Types that files may be bound to, and a func that can read the content from
	// that type.
	type fh struct {
		filename string
		content  []byte
	}
	type test struct {
		key string
		fhs []fh
	}

	var expectedFiles = map[string][]fh{
		"file1": {fh{"test.txt", []byte("content1")}},
		"file2": {fh{"test.txt", []byte("content2")}, fh{"favicon.ico", []byte("xyz")}},
	}

	var (
		file1FileHeader *multipart.FileHeader
		file1File       *os.File
		file1Bytes      []byte
		file1Reader     io.Reader
		file1ReadSeeker io.ReadSeeker
	)
	recordErrors(t,
		binder.Field(&file1FileHeader, "file1"),
		binder.Field(&file1File, "file1"),
		binder.Field(&file1Bytes, "file1"),
		binder.Field(&file1Reader, "file1"),
		binder.Field(&file1ReadSeeker, "file1"),
	)

	// The only one of these that keeps the filename from the form around is
	// multipart.FileHeader.
	if file1FileHeader.Filename != expectedFiles["file1"][0].filename {
		t.Errorf("didn't bind name. got %q", file1FileHeader.Filename)
	}

	// Check the content
	for _, reader := range []io.Reader{
		openFileHeader(t, file1FileHeader),
		file1File,
		bytes.NewReader(file1Bytes),
		file1Reader,
		file1ReadSeeker,
	} {
		actual, _ := ioutil.ReadAll(reader)
		if !bytes.Equal(actual, expectedFiles["file1"][0].content) {
			t.Error("got different bytes")
		}
	}

	// Check binding a slice of files.
	var (
		file2FileHeader []*multipart.FileHeader
		file2File       []*os.File
		file2Bytes      [][]byte
		file2Reader     []io.Reader
		file2ReadSeeker []io.ReadSeeker
	)
	recordErrors(t,
		binder.Field(&file2FileHeader, "file2"),
		binder.Field(&file2File, "file2"),
		binder.Field(&file2Bytes, "file2"),
		binder.Field(&file2Reader, "file2"),
		binder.Field(&file2ReadSeeker, "file2"),
	)

	if len(file2FileHeader) != 2 {
		t.Error("didn't return 2 files")
		return
	}
	if file2FileHeader[0].Filename != expectedFiles["file2"][0].filename {
		t.Errorf("didn't bind name. got %q", file2FileHeader[0].Filename)
	}
	if file2FileHeader[1].Filename != expectedFiles["file2"][1].filename {
		t.Errorf("didn't bind name. got %q", file2FileHeader[1].Filename)
	}

	// Check the content
	for i, fh := range expectedFiles["file2"] {
		for _, reader := range []io.Reader{
			openFileHeader(t, file2FileHeader[i]),
			file2File[i],
			bytes.NewReader(file2Bytes[i]),
			file2Reader[i],
			file2ReadSeeker[i],
		} {
			actual, _ := ioutil.ReadAll(reader)
			if !bytes.Equal(actual, fh.content) {
				t.Error("got different bytes")
			}
		}
	}
}

func openFileHeader(t *testing.T, fh *multipart.FileHeader) multipart.File {
	var r, err = fh.Open()
	if err != nil {
		t.Error(err)
	}
	return r
}
