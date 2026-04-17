window.BENCHMARK_DATA = {
  "lastUpdate": 1776412046427,
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
      }
    ]
  }
}