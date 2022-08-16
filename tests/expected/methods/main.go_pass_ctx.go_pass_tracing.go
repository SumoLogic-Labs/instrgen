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
	"context"

	"github.com/pdelewski/autotel/rtlib"
	otel "go.opentelemetry.io/otel"
)

type element struct {
}

type driver struct {
	e element
}

type i interface {
	foo(__tracing_ctx context.Context, p int) int
}

type impl struct {
}

func (i impl) foo(__tracing_ctx context.Context, p int) int {
	__child_tracing_ctx, span := otel.Tracer("foo").Start(__tracing_ctx, "foo")
	_ = __child_tracing_ctx
	defer span.End()
	return 5
}

func foo(__tracing_ctx context.Context, p int) int {
	__child_tracing_ctx, span := otel.Tracer("foo").Start(__tracing_ctx, "foo")
	_ = __child_tracing_ctx
	defer span.End()
	return 1
}

func (d driver) process(__tracing_ctx context.Context, a int) {
	__child_tracing_ctx, span := otel.Tracer("process").Start(__tracing_ctx, "process")
	_ = __child_tracing_ctx
	defer span.End()
}

func (e element) get(__tracing_ctx context.Context, a int) {
	__child_tracing_ctx, span := otel.Tracer("get").Start(__tracing_ctx, "get")
	_ = __child_tracing_ctx
	defer span.End()
}

func main() {
	__child_tracing_ctx := context.TODO()
	_ = __child_tracing_ctx
	ts := rtlib.NewTracingState()
	defer rtlib.Shutdown(ts)
	otel.SetTracerProvider(ts.Tp)
	ctx := context.Background()
	__child_tracing_ctx, span := otel.Tracer("main").Start(ctx, "main")
	defer span.End()
	rtlib.AutotelEntryPoint__()
	d := driver{}
	d.process(__child_tracing_ctx, 10)
	d.e.get(__child_tracing_ctx, 5)
	var in i
	in = impl{}
	in.foo(__child_tracing_ctx, 10)
}
