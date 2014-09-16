go-metrics
==========

High-performance metrics package for Go. 

Supported metric types are
- Gauge: a single instantaneous value
- Counter: a single integer value that may be incremented or decremented
- Meter: a single integer value and its derivatives over time.
- Distribution: stores a sample of a data set and computes statistics like mean, median, percentiles, etc. 

Statistics are computed as data is added. All operations except retrieving a
distribution's sample are O(log n) or faster.

Metrics may be registered under names. The 'dashboard' package provides an HTTP server that exports collected data and statistics in JSON and graphical formats. 
