window.BENCHMARK_DATA = {
  "lastUpdate": 1777799558722,
  "repoUrl": "https://github.com/txabman42/opentelemetry-go-compile-instrumentation",
  "entries": {
    "otelc Compile-Time Benchmarks": [
      {
        "commit": {
          "author": {
            "email": "39015378+txabman42@users.noreply.github.com",
            "name": "Xabier Martinez",
            "username": "txabman42"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "965787b615717cb455c104a0ff5e4bf1d3b1763e",
          "message": "Merge pull request #42 from txabman42/xabier/fix\n\nchore: fix",
          "timestamp": "2026-04-17T09:38:15+02:00",
          "tree_id": "9cf59a8498d4568bdba30e001dbc1609aebfde10",
          "url": "https://github.com/txabman42/opentelemetry-go-compile-instrumentation/commit/965787b615717cb455c104a0ff5e4bf1d3b1763e"
        },
        "date": 1776412045433,
        "tool": "customSmallerIsBetter",
        "benches": [
          {
            "name": "baseline / otelc compile time",
            "value": 4.59,
            "range": "± 0.018",
            "unit": "s",
            "extra": "plain=3.622s (±0.022) overhead=+26.7%"
          },
          {
            "name": "baseline / plain compile time",
            "value": 3.622,
            "range": "± 0.022",
            "unit": "s"
          },
          {
            "name": "baseline / overhead",
            "value": 26.732,
            "unit": "%"
          },
          {
            "name": "largeidle / otelc compile time",
            "value": 14.743,
            "range": "± 0.022",
            "unit": "s",
            "extra": "plain=13.193s (±0.071) overhead=+11.7%"
          },
          {
            "name": "largeidle / plain compile time",
            "value": 13.193,
            "range": "± 0.071",
            "unit": "s"
          },
          {
            "name": "largeidle / overhead",
            "value": 11.746,
            "unit": "%"
          },
          {
            "name": "multi / otelc compile time",
            "value": 24.459,
            "range": "± 0.032",
            "unit": "s",
            "extra": "plain=15.578s (±0.019) overhead=+57.0%"
          },
          {
            "name": "multi / plain compile time",
            "value": 15.578,
            "range": "± 0.019",
            "unit": "s"
          },
          {
            "name": "multi / overhead",
            "value": 57.002,
            "unit": "%"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "49699333+dependabot[bot]@users.noreply.github.com",
            "name": "dependabot[bot]",
            "username": "dependabot[bot]"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "8935b001a81cfccd92591c57f0b7891b2f6849d0",
          "message": "chore(deps): bump go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp from 0.14.0 to 0.19.0 in /pkg/instrumentation/shared (#445)\n\nBumps\n[go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp](https://github.com/open-telemetry/opentelemetry-go)\nfrom 0.14.0 to 0.19.0.\n<details>\n<summary>Release notes</summary>\n<p><em>Sourced from <a\nhref=\"https://github.com/open-telemetry/opentelemetry-go/releases\">go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp's\nreleases</a>.</em></p>\n<blockquote>\n<h2>Release v0.19.0</h2>\n<h3>Added</h3>\n<ul>\n<li>Added <code>Marshaler</code> config option to <code>otlphttp</code>\nto enable otlp over json or protobufs. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1586\">#1586</a>)</li>\n<li>A <code>ForceFlush</code> method to the\n<code>&quot;go.opentelemetry.io/otel/sdk/trace&quot;.TracerProvider</code>\nto flush all registered <code>SpanProcessor</code>s. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1608\">#1608</a>)</li>\n<li>Added <code>WithSampler</code> and <code>WithSpanLimits</code> to\ntracer provider. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1633\">#1633</a>,\n<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1702\">#1702</a>)</li>\n<li><code>&quot;go.opentelemetry.io/otel/trace&quot;.SpanContext</code>\nnow has a <code>remote</code> property, and <code>IsRemote()</code>\npredicate, that is true when the <code>SpanContext</code> has been\nextracted from remote context data. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1701\">#1701</a>)</li>\n<li>A <code>Valid</code> method to the\n<code>&quot;go.opentelemetry.io/otel/attribute&quot;.KeyValue</code>\ntype. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1703\">#1703</a>)</li>\n</ul>\n<h3>Changed</h3>\n<ul>\n<li><code>trace.SpanContext</code> is now immutable and has no exported\nfields. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1573\">#1573</a>)\n<ul>\n<li><code>trace.NewSpanContext()</code> can be used in conjunction with\nthe <code>trace.SpanContextConfig</code> struct to initialize a new\n<code>SpanContext</code> where all values are known.</li>\n</ul>\n</li>\n<li>Update the <code>ForceFlush</code> method signature to the\n<code>&quot;go.opentelemetry.io/otel/sdk/trace&quot;.SpanProcessor</code>\nto accept a <code>context.Context</code> and return an error. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1608\">#1608</a>)</li>\n<li>Update the <code>Shutdown</code> method to the\n<code>&quot;go.opentelemetry.io/otel/sdk/trace&quot;.TracerProvider</code>\nreturn an error on shutdown failure. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1608\">#1608</a>)</li>\n<li>The SimpleSpanProcessor will now shut down the enclosed\n<code>SpanExporter</code> and gracefully ignore subsequent calls to\n<code>OnEnd</code> after <code>Shutdown</code> is called. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1612\">#1612</a>)</li>\n\n<li><code>&quot;go.opentelemetry.io/sdk/metric/controller.basic&quot;.WithPusher</code>\nis replaced with <code>WithExporter</code> to provide consistent naming\nacross project. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1656\">#1656</a>)</li>\n<li>Added non-empty string check for trace <code>Attribute</code> keys.\n(<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1659\">#1659</a>)</li>\n<li>Add <code>description</code> to SpanStatus only when\n<code>StatusCode</code> is set to error. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1662\">#1662</a>)</li>\n<li>Jaeger exporter falls back to <code>resource.Default</code>'s\n<code>service.name</code> if the exported Span does not have one. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1673\">#1673</a>)</li>\n<li>Jaeger exporter populates Jaeger's Span Process from Resource. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1673\">#1673</a>)</li>\n<li>Renamed the <code>LabelSet</code> method of\n<code>&quot;go.opentelemetry.io/otel/sdk/resource&quot;.Resource</code>\nto <code>Set</code>. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1692\">#1692</a>)</li>\n<li>Changed <code>WithSDK</code> to <code>WithSDKOptions</code> to\naccept variadic arguments of <code>TracerProviderOption</code> type in\n<code>go.opentelemetry.io/otel/exporters/trace/jaeger</code> package.\n(<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1693\">#1693</a>)</li>\n<li>Changed <code>WithSDK</code> to <code>WithSDKOptions</code> to\naccept variadic arguments of <code>TracerProviderOption</code> type in\n<code>go.opentelemetry.io/otel/exporters/trace/zipkin</code> package.\n(<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1693\">#1693</a>)</li>\n\n<li><code>&quot;go.opentelemetry.io/otel/sdk/resource&quot;.NewWithAttributes</code>\nwill now drop any invalid attributes passed. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1703\">#1703</a>)</li>\n\n<li><code>&quot;go.opentelemetry.io/otel/sdk/resource&quot;.StringDetector</code>\nwill now error if the produced attribute is invalid. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1703\">#1703</a>)</li>\n</ul>\n<h3>Removed</h3>\n<ul>\n<li>Removed <code>serviceName</code> parameter from Zipkin exporter and\nuses resource instead. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1549\">#1549</a>)</li>\n<li>Removed <code>WithConfig</code> from tracer provider to avoid\noverriding configuration. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1633\">#1633</a>)</li>\n<li>Removed the exported <code>SimpleSpanProcessor</code> and\n<code>BatchSpanProcessor</code> structs.\nThese are now returned as a SpanProcessor interface from their\nrespective constructors. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1638\">#1638</a>)</li>\n<li>Removed <code>WithRecord()</code> from <code>trace.SpanOption</code>\nwhen creating a span. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1660\">#1660</a>)</li>\n<li>Removed setting status to <code>Error</code> while recording an\nerror as a span event in <code>RecordError</code>. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1663\">#1663</a>)</li>\n<li>Removed <code>jaeger.WithProcess</code> configuration option. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1673\">#1673</a>)</li>\n<li>Removed <code>ApplyConfig</code> method from\n<code>&quot;go.opentelemetry.io/otel/sdk/trace&quot;.TracerProvider</code>\nand the now unneeded <code>Config</code> struct. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1693\">#1693</a>)</li>\n</ul>\n<h3>Fixed</h3>\n<ul>\n<li>Jaeger Exporter: Ensure mapping between OTEL and Jaeger span data\ncomplies with the specification. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1626\">#1626</a>)</li>\n<li><code>SamplingResult.TraceState</code> is correctly propagated to a\nnewly created span's <code>SpanContext</code>. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1655\">#1655</a>)</li>\n<li>The <code>otel-collector</code> example now correctly flushes metric\nevents prior to shutting down the exporter. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1678\">#1678</a>)</li>\n<li>Do not set span status message in\n<code>SpanStatusFromHTTPStatusCode</code> if it can be inferred from\n<code>http.status_code</code>. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1681\">#1681</a>)</li>\n<li>Synchronization issues in global trace delegate implementation. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1686\">#1686</a>)</li>\n<li>Reduced excess memory usage by global <code>TracerProvider</code>.\n(<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1687\">#1687</a>)</li>\n</ul>\n<hr />\n<h2>Raw changes made between v0.18.0 and v0.19.0</h2>\n<!-- raw HTML omitted -->\n</blockquote>\n<p>... (truncated)</p>\n</details>\n<details>\n<summary>Changelog</summary>\n<p><em>Sourced from <a\nhref=\"https://github.com/open-telemetry/opentelemetry-go/blob/main/CHANGELOG.md\">go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp's\nchangelog</a>.</em></p>\n<blockquote>\n<h2>[1.43.0/0.65.0/0.19.0] 2026-04-02</h2>\n<h3>Added</h3>\n<ul>\n<li>Add <code>IsRandom</code> and <code>WithRandom</code> on\n<code>TraceFlags</code>, and <code>IsRandom</code> on\n<code>SpanContext</code> in <code>go.opentelemetry.io/otel/trace</code>\nfor <a\nhref=\"https://www.w3.org/TR/trace-context-2/#random-trace-id-flag\">W3C\nTrace Context Level 2 Random Trace ID Flag</a> support. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8012\">#8012</a>)</li>\n<li>Add service detection with <code>WithService</code> in\n<code>go.opentelemetry.io/otel/sdk/resource</code>. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/7642\">#7642</a>)</li>\n<li>Add <code>DefaultWithContext</code> and\n<code>EnvironmentWithContext</code> in\n<code>go.opentelemetry.io/otel/sdk/resource</code> to support plumbing\n<code>context.Context</code> through default and environment detectors.\n(<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8051\">#8051</a>)</li>\n<li>Support attributes with empty value (<code>attribute.EMPTY</code>)\nin\n<code>go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc</code>.\n(<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8038\">#8038</a>)</li>\n<li>Support attributes with empty value (<code>attribute.EMPTY</code>)\nin\n<code>go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc</code>.\n(<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8038\">#8038</a>)</li>\n<li>Support attributes with empty value (<code>attribute.EMPTY</code>)\nin\n<code>go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc</code>.\n(<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8038\">#8038</a>)</li>\n<li>Support attributes with empty value (<code>attribute.EMPTY</code>)\nin\n<code>go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp</code>.\n(<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8038\">#8038</a>)</li>\n<li>Support attributes with empty value (<code>attribute.EMPTY</code>)\nin\n<code>go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp</code>.\n(<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8038\">#8038</a>)</li>\n<li>Support attributes with empty value (<code>attribute.EMPTY</code>)\nin\n<code>go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp</code>.\n(<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8038\">#8038</a>)</li>\n<li>Support attributes with empty value (<code>attribute.EMPTY</code>)\nin\n<code>go.opentelemetry.io/otel/sdk/metric/metricdata/metricdatatest</code>.\n(<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8038\">#8038</a>)</li>\n<li>Add support for per-series start time tracking for cumulative\nmetrics in <code>go.opentelemetry.io/otel/sdk/metric</code>.\nSet <code>OTEL_GO_X_PER_SERIES_START_TIMESTAMPS=true</code> to enable.\n(<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8060\">#8060</a>)</li>\n<li>Add <code>WithCardinalityLimitSelector</code> for metric reader for\nconfiguring cardinality limits specific to the instrument kind. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/7855\">#7855</a>)</li>\n</ul>\n<h3>Changed</h3>\n<ul>\n<li>Introduce the <code>EMPTY</code> Type in\n<code>go.opentelemetry.io/otel/attribute</code> to reflect that an empty\nvalue is now a valid value, with <code>INVALID</code> remaining as a\ndeprecated alias of <code>EMPTY</code>. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8038\">#8038</a>)</li>\n<li>Improve slice handling in\n<code>go.opentelemetry.io/otel/attribute</code> to optimize short slice\nvalues with fixed-size fast paths. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8039\">#8039</a>)</li>\n<li>Improve performance of span metric recording in\n<code>go.opentelemetry.io/otel/sdk/trace</code> by returning early if\nself-observability is not enabled. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8067\">#8067</a>)</li>\n<li>Improve formatting of metric data diffs in\n<code>go.opentelemetry.io/otel/sdk/metric/metricdata/metricdatatest</code>.\n(<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8073\">#8073</a>)</li>\n</ul>\n<h3>Deprecated</h3>\n<ul>\n<li>Deprecate <code>INVALID</code> in\n<code>go.opentelemetry.io/otel/attribute</code>. Use <code>EMPTY</code>\ninstead. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8038\">#8038</a>)</li>\n</ul>\n<h3>Fixed</h3>\n<ul>\n<li>Return spec-compliant <code>TraceIdRatioBased</code> description.\nThis is a breaking behavioral change, but it is necessary to\nmake the implementation <a\nhref=\"https://opentelemetry.io/docs/specs/otel/trace/sdk/#traceidratiobased\">spec-compliant</a>.\n(<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8027\">#8027</a>)</li>\n<li>Fix a race condition in\n<code>go.opentelemetry.io/otel/sdk/metric</code> where the lastvalue\naggregation could collect the value 0 even when no zero-value\nmeasurements were recorded. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8056\">#8056</a>)</li>\n<li>Limit HTTP response body to 4 MiB in\n<code>go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp</code>\nto mitigate excessive memory usage caused by a misconfigured or\nmalicious server.\nResponses exceeding the limit are treated as non-retryable errors. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8108\">#8108</a>)</li>\n<li>Limit HTTP response body to 4 MiB in\n<code>go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp</code>\nto mitigate excessive memory usage caused by a misconfigured or\nmalicious server.\nResponses exceeding the limit are treated as non-retryable errors. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8108\">#8108</a>)</li>\n<li>Limit HTTP response body to 4 MiB in\n<code>go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp</code>\nto mitigate excessive memory usage caused by a misconfigured or\nmalicious server.\nResponses exceeding the limit are treated as non-retryable errors. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8108\">#8108</a>)</li>\n<li><code>WithHostID</code> detector in\n<code>go.opentelemetry.io/otel/sdk/resource</code> to use full path for\n<code>kenv</code> command on BSD. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8113\">#8113</a>)</li>\n<li>Fix missing <code>request.GetBody</code> in\n<code>go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp</code>\nto correctly handle HTTP2 GOAWAY frame. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/8096\">#8096</a>)</li>\n</ul>\n<h2>[1.42.0/0.64.0/0.18.0/0.0.16] 2026-03-06</h2>\n<h3>Added</h3>\n<ul>\n<li>Add <code>go.opentelemetry.io/otel/semconv/v1.40.0</code> package.\nThe package contains semantic conventions from the <code>v1.40.0</code>\nversion of the OpenTelemetry Semantic Conventions.\nSee the <a\nhref=\"https://github.com/open-telemetry/opentelemetry-go/blob/main/semconv/v1.40.0/MIGRATION.md\">migration\ndocumentation</a> for information on how to upgrade from\n<code>go.opentelemetry.io/otel/semconv/v1.39.0</code>. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/7985\">#7985</a>)</li>\n</ul>\n<!-- raw HTML omitted -->\n</blockquote>\n<p>... (truncated)</p>\n</details>\n<details>\n<summary>Commits</summary>\n<ul>\n<li><a\nhref=\"https://github.com/open-telemetry/opentelemetry-go/commit/2b4fa9681bd0c69574aaa879039382002b220204\"><code>2b4fa96</code></a>\nRelease v0.19.0 (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1710\">#1710</a>)</li>\n<li><a\nhref=\"https://github.com/open-telemetry/opentelemetry-go/commit/4beb70416e1272c578edfe1d5f88a3a2236da178\"><code>4beb704</code></a>\nsdk/trace: removing ApplyConfig and Config (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1693\">#1693</a>)</li>\n<li><a\nhref=\"https://github.com/open-telemetry/opentelemetry-go/commit/1d42be1601e2d9bbd1101780759520e3f3960a29\"><code>1d42be1</code></a>\nRename WithDefaultSampler TracerProvider option to WithSampler and\nupdate doc...</li>\n<li><a\nhref=\"https://github.com/open-telemetry/opentelemetry-go/commit/860d5d86e7ace12bf2b2ca8e437d2d4fc68a6913\"><code>860d5d8</code></a>\nAdd flag to determine whether SpanContext is remote (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1701\">#1701</a>)</li>\n<li><a\nhref=\"https://github.com/open-telemetry/opentelemetry-go/commit/0fe65e6bd2b3fad00289427e0bac1974086d4326\"><code>0fe65e6</code></a>\nComply with OpenTelemetry attributes specification (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1703\">#1703</a>)</li>\n<li><a\nhref=\"https://github.com/open-telemetry/opentelemetry-go/commit/888843519dae308f165d1d20c095bb6352baeb52\"><code>8888435</code></a>\nBump google.golang.org/api from 0.40.0 to 0.41.0 in\n/exporters/trace/jaeger (...</li>\n<li><a\nhref=\"https://github.com/open-telemetry/opentelemetry-go/commit/345f264a137ed7162c30d14dd4739b5b72f76537\"><code>345f264</code></a>\nbreaking(zipkin): removes servicName from zipkin exporter. (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1697\">#1697</a>)</li>\n<li><a\nhref=\"https://github.com/open-telemetry/opentelemetry-go/commit/62cbf0f240112813105d7056506496b59740e0c2\"><code>62cbf0f</code></a>\nPopulate Jaeger's Span.Process from Resource (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1673\">#1673</a>)</li>\n<li><a\nhref=\"https://github.com/open-telemetry/opentelemetry-go/commit/28eaaa9a919d03227856d83e2149b85f78d57775\"><code>28eaaa9</code></a>\nAdd a test to prove the Tracer is safe for concurrent calls (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1665\">#1665</a>)</li>\n<li><a\nhref=\"https://github.com/open-telemetry/opentelemetry-go/commit/8b1be11a549eefb6efeda2f940cbda70b3c3d08d\"><code>8b1be11</code></a>\nRename resource pkg label vars and methods (<a\nhref=\"https://redirect.github.com/open-telemetry/opentelemetry-go/issues/1692\">#1692</a>)</li>\n<li>Additional commits viewable in <a\nhref=\"https://github.com/open-telemetry/opentelemetry-go/compare/v0.14.0...v0.19.0\">compare\nview</a></li>\n</ul>\n</details>\n<br />\n\n\n[![Dependabot compatibility\nscore](https://dependabot-badges.githubapp.com/badges/compatibility_score?dependency-name=go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp&package-manager=go_modules&previous-version=0.14.0&new-version=0.19.0)](https://docs.github.com/en/github/managing-security-vulnerabilities/about-dependabot-security-updates#about-compatibility-scores)\n\nDependabot will resolve any conflicts with this PR as long as you don't\nalter it yourself. You can also trigger a rebase manually by commenting\n`@dependabot rebase`.\n\n[//]: # (dependabot-automerge-start)\n[//]: # (dependabot-automerge-end)\n\n---\n\n<details>\n<summary>Dependabot commands and options</summary>\n<br />\n\nYou can trigger Dependabot actions by commenting on this PR:\n- `@dependabot rebase` will rebase this PR\n- `@dependabot recreate` will recreate this PR, overwriting any edits\nthat have been made to it\n- `@dependabot show <dependency name> ignore conditions` will show all\nof the ignore conditions of the specified dependency\n- `@dependabot ignore this major version` will close this PR and stop\nDependabot creating any more for this major version (unless you reopen\nthe PR or upgrade to it yourself)\n- `@dependabot ignore this minor version` will close this PR and stop\nDependabot creating any more for this minor version (unless you reopen\nthe PR or upgrade to it yourself)\n- `@dependabot ignore this dependency` will close this PR and stop\nDependabot creating any more for this dependency (unless you reopen the\nPR or upgrade to it yourself)\nYou can disable automated security fix PRs for this repo from the\n[Security Alerts\npage](https://github.com/open-telemetry/opentelemetry-go-compile-instrumentation/network/alerts).\n\n</details>\n\nSigned-off-by: dependabot[bot] <support@github.com>\nCo-authored-by: dependabot[bot] <49699333+dependabot[bot]@users.noreply.github.com>",
          "timestamp": "2026-04-28T17:21:32+08:00",
          "tree_id": "3c8a12137692b95d4323a44924b6002ecff66bed",
          "url": "https://github.com/txabman42/opentelemetry-go-compile-instrumentation/commit/8935b001a81cfccd92591c57f0b7891b2f6849d0"
        },
        "date": 1777799558251,
        "tool": "customSmallerIsBetter",
        "benches": [
          {
            "name": "baseline / otelc compile time",
            "value": 4.433,
            "range": "± 0.007",
            "unit": "s",
            "extra": "plain=3.494s (±0.011) overhead=+26.9%"
          },
          {
            "name": "baseline / plain compile time",
            "value": 3.494,
            "range": "± 0.011",
            "unit": "s"
          },
          {
            "name": "baseline / overhead",
            "value": 26.863,
            "unit": "%"
          },
          {
            "name": "largeidle / otelc compile time",
            "value": 14.041,
            "range": "± 0.031",
            "unit": "s",
            "extra": "plain=12.572s (±0.008) overhead=+11.7%"
          },
          {
            "name": "largeidle / plain compile time",
            "value": 12.572,
            "range": "± 0.008",
            "unit": "s"
          },
          {
            "name": "largeidle / overhead",
            "value": 11.678,
            "unit": "%"
          },
          {
            "name": "multi / otelc compile time",
            "value": 24.06,
            "range": "± 0.043",
            "unit": "s",
            "extra": "plain=15.016s (±0.038) overhead=+60.2%"
          },
          {
            "name": "multi / plain compile time",
            "value": 15.016,
            "range": "± 0.038",
            "unit": "s"
          },
          {
            "name": "multi / overhead",
            "value": 60.233,
            "unit": "%"
          }
        ]
      }
    ]
  }
}