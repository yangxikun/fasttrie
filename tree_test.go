package fasttrie

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/savsgio/gotils"
	"github.com/valyala/bytebufferpool"
)

func generateValue() string {
	hex := gotils.RandBytes(make([]byte, 10))
	return fmt.Sprintf("%x", hex)
}

func testValueAndParams(
	t *testing.T, tree *Tree, reqPath string, value interface{}, wantTSR bool, params map[string]string,
) {
	for _, parsedParams := range []map[string]string{make(map[string]string), nil} {

		h, tsr := tree.Get(reqPath, parsedParams)
		if value != h {
			t.Errorf("Path '%s' value == %p, want %p", reqPath, h, value)
		}

		if wantTSR != tsr {
			t.Errorf("Path '%s' tsr == %v, want %v", reqPath, tsr, wantTSR)
		}

		if parsedParams != nil {
			if params == nil {
				params = make(map[string]string)
			}
			if !reflect.DeepEqual(parsedParams, params) {
				t.Errorf("Path '%s' User values == %v, want %v", reqPath, parsedParams, params)
			}
		}
	}
}

func Test_Tree(t *testing.T) {
	type args struct {
		path    string
		reqPath string
		value   interface{}
	}

	type want struct {
		tsr    bool
		params map[string]string
	}

	tests := []struct {
		args args
		want want
	}{
		{
			args: args{
				path:    "/users/{name}",
				reqPath: "/users/atreugo",
				value:   generateValue(),
			},
			want: want{
				params: map[string]string{
					"name": "atreugo",
				},
			},
		},
		{
			args: args{
				path:    "/users",
				reqPath: "/users",
				value:   generateValue(),
			},
			want: want{
				params: nil,
			},
		},
		{
			args: args{
				path:    "/user/",
				reqPath: "/user",
				value:   generateValue(),
			},
			want: want{
				tsr:    true,
				params: nil,
			},
		},
		{
			args: args{
				path:    "/",
				reqPath: "/",
				value:   generateValue(),
			},
			want: want{
				params: nil,
			},
		},
		{
			args: args{
				path:    "/users/{name}/jobs",
				reqPath: "/users/atreugo/jobs",
				value:   generateValue(),
			},
			want: want{
				params: map[string]string{
					"name": "atreugo",
				},
			},
		},
		{
			args: args{
				path:    "/users/admin",
				reqPath: "/users/admin",
				value:   generateValue(),
			},
			want: want{
				params: nil,
			},
		},
		{
			args: args{
				path:    "/users/{status}/proc",
				reqPath: "/users/active/proc",
				value:   generateValue(),
			},
			want: want{
				params: map[string]string{
					"status": "active",
				},
			},
		},
		{
			args: args{
				path:    "/static/{filepath:*}",
				reqPath: "/static/assets/js/main.js",
				value:   generateValue(),
			},
			want: want{
				params: map[string]string{
					"filepath": "assets/js/main.js",
				},
			},
		},
	}

	tree := New()

	for _, test := range tests {
		tree.Add(test.args.path, test.args.value)
	}

	for _, test := range tests {
		wantValue := test.args.value
		if test.want.tsr {
			wantValue = nil
		}

		testValueAndParams(t, tree, test.args.reqPath, wantValue, test.want.tsr, test.want.params)
	}

	filepathValue := generateValue()

	tree.Add("/{filepath:*}", filepathValue)

	testValueAndParams(t, tree, "/js/main.js", filepathValue, false, map[string]string{
		"filepath": "js/main.js",
	})
}

func Test_Get(t *testing.T) {
	value := generateValue()

	tree := New()
	tree.Add("/api/", value)

	testValueAndParams(t, tree, "/api", nil, true, nil)
	testValueAndParams(t, tree, "/api/", value, false, nil)
	testValueAndParams(t, tree, "/notfound", nil, false, nil)

	tree = New()
	tree.Add("/api", value)

	testValueAndParams(t, tree, "/api", value, false, nil)
	testValueAndParams(t, tree, "/api/", nil, true, nil)
	testValueAndParams(t, tree, "/notfound", nil, false, nil)
}

func Test_AddWithParam(t *testing.T) {
	value := generateValue()

	tree := New()
	tree.Add("/test", value)
	tree.Add("/api/prefix{version:V[0-9]}_{name:[a-z]+}_sufix/files", value)
	tree.Add("/api/prefix{version:V[0-9]}_{name:[a-z]+}_sufix/data", value)
	tree.Add("/api/prefix/files", value)
	tree.Add("/prefix{name:[a-z]+}suffix/data", value)
	tree.Add("/prefix{name:[a-z]+}/data", value)
	tree.Add("/api/{file}.json", value)

	testValueAndParams(t, tree, "/api/prefixV1_atreugo_sufix/files", value, false, map[string]string{
		"version": "V1", "name": "atreugo",
	})
	testValueAndParams(t, tree, "/api/prefixV1_atreugo_sufix/data", value, false, map[string]string{
		"version": "V1", "name": "atreugo",
	})
	testValueAndParams(t, tree, "/prefixatreugosuffix/data", value, false, map[string]string{
		"name": "atreugo",
	})
	testValueAndParams(t, tree, "/prefixatreugo/data", value, false, map[string]string{
		"name": "atreugo",
	})
	testValueAndParams(t, tree, "/api/name.json", value, false, map[string]string{
		"file": "name",
	})

	// Not found
	testValueAndParams(t, tree, "/api/prefixV1_1111_sufix/fake", nil, false, nil)

}

func Test_TreeRootWildcard(t *testing.T) {
	value := generateValue()

	tree := New()
	tree.Add("/{filepath:*}", value)

	testValueAndParams(t, tree, "/", value, false, map[string]string{
		"filepath": "/",
	})
}

func Test_TreeNilValue(t *testing.T) {
	const panicMsg = "nil value"

	tree := New()

	err := catchPanic(func() {
		tree.Add("/", nil)
	})

	if err == nil {
		t.Fatal("Expected panic")
	}

	if err != nil && panicMsg != fmt.Sprint(err) {
		t.Errorf("Invalid conflict error text (%v)", err)
	}
}

func Test_TreeMutable(t *testing.T) {
	routes := []string{
		"/",
		"/api/{version}",
		"/{filepath:*}",
		"/user{user:a-Z+}",
	}

	value := generateValue()
	tree := New()

	for _, route := range routes {
		tree.Add(route, value)

		err := catchPanic(func() {
			tree.Add(route, value)
		})

		if err == nil {
			t.Errorf("Route '%s' - Expected panic", route)
		}
	}

	tree.Mutable = true

	for _, route := range routes {
		err := catchPanic(func() {
			tree.Add(route, value)
		})

		if err != nil {
			t.Errorf("Route '%s' - Unexpected panic: %v", route, err)
		}
	}
}

func Benchmark_Get(b *testing.B) {
	value := struct{}{}

	tree := New()

	// for i := 0; i < 3000; i++ {
	// 	tree.Add(
	// 		fmt.Sprintf("/%s", gotils.RandBytes(make([]byte, 15))), value,
	// 	)
	// }

	tree.Add("/", value)
	tree.Add("/plaintext", value)
	tree.Add("/json", value)
	tree.Add("/fortune", value)
	tree.Add("/fortune-quick", value)
	tree.Add("/db", value)
	tree.Add("/queries", value)
	tree.Add("/update", value)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tree.Get("/update", nil)
	}
}

func Benchmark_GetWithRegex(b *testing.B) {
	tree := New()

	tree.Add("/api/{version:v[0-9]}/data", struct{}{})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tree.Get("/api/v1/data", nil)
	}
}

func Benchmark_GetWithParams(b *testing.B) {
	tree := New()

	tree.Add("/api/{version}/data", struct{}{})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tree.Get("/api/v1/data", nil)
	}
}

func Benchmark_FindCaseInsensitivePath(b *testing.B) {
	tree := New()
	buf := bytebufferpool.Get()

	tree.Add("/endpoint", struct{}{})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tree.FindCaseInsensitivePath("/ENdpOiNT", false, buf)
		buf.Reset()
	}
}
