// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/pdelewski/autotel/rtlib"
	__atel_otel "go.opentelemetry.io/otel"
	__atel_context "context"
)

type element struct {
}

type driver struct {
	e element
}

type i interface {
	foo(__atel_tracing_ctx __atel_context.Context, p int) int
}

type impl struct {
}

func (i impl) foo(__atel_tracing_ctx __atel_context.Context, p int) int {
	__atel_child_tracing_ctx, __atel_span := __atel_otel.Tracer("foo").Start(__atel_tracing_ctx, "foo")
	_ = __atel_child_tracing_ctx
	defer __atel_span.End()
	return 5
}

func foo(p int) int {
	return 1
}

func (d driver) process(__atel_tracing_ctx __atel_context.Context, a int) {
	__atel_child_tracing_ctx, __atel_span := __atel_otel.Tracer("process").Start(__atel_tracing_ctx, "process")
	_ = __atel_child_tracing_ctx
	defer __atel_span.End()

}

func (e element) get(__atel_tracing_ctx __atel_context.Context, a int) {
	__atel_child_tracing_ctx, __atel_span := __atel_otel.Tracer("get").Start(__atel_tracing_ctx, "get")
	_ = __atel_child_tracing_ctx
	defer __atel_span.End()

}

func main() {
	__atel_ts := rtlib.NewTracingState()
	defer rtlib.Shutdown(__atel_ts)
	__atel_otel.SetTracerProvider(__atel_ts.Tp)
	__atel_ctx := __atel_context.Background()
	__atel_child_tracing_ctx, __atel_span := __atel_otel.Tracer("main").Start(__atel_ctx, "main")
	_ = __atel_child_tracing_ctx
	defer __atel_span.End()

	rtlib.AutotelEntryPoint__()
	d := driver{}
	d.process(__atel_child_tracing_ctx, 10)
	d.e.get(__atel_child_tracing_ctx, 5)
	var in i
	in = impl{}
	in.foo(__atel_child_tracing_ctx, 10)
}
