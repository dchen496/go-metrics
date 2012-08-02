package dashboard
const index_html = `<html>
  <head>
    <title>{{.Name}} - Metrics Dashboard</title>
    <style>
      * {
        font-family: sans-serif;
        font-size: 10pt;
      }

      tr:nth-child(even){
        background-color:#e0e0e0
      }

      table {
        border-collapse: collapse;
      }

      td,th {
        border-left: 1px solid black;
        padding-left: 3px;
        padding-right: 3px;
      }

      th:first-child,
      td:first-child{
        border-left: 0;
      }

      div.horizon > canvas {
        display: block;
      }
      div.horizon > span.value {
        position: relative;
        left: 20px;
      }

      div.graph{
        float: left;
        width: 400px;
      }

      canvas#heatmap{
        margin-top: 5px;
        margin-bottom: 5px;
      }

    </style>
  </head>
  <body>
    <h3>{{.Name}} - Metrics Dashboard</h3>
    <p>Usage: Click on a checkbox to show/hide the metric.
    Click on a table cell to show its corresponding graph.
    Not all table cells have graphs. Clicking on a Distribution's name
    will show its heatmap. Clicking on its count will show a smoothed
    histogram.
    </p>
    <button id="hidehides"></button>
    <div id="hides" style="display:none"></div>
    <div id="metrics">
      <h3>Counters</h3>
      <table id="counters">
        <tr id="header"><th>Name</th><th>Value</th><th>Updated</th></tr>
      </table>
      <h3>Distributions</h3>
      <table id="distributions">
        <tr id="header">
          <th>Name</th>
          <th>Count</th>
          <th>Mean</th>
          <th>Variance</th>
          <th>StdDev</th>
          <th>Skewness</th>
          <th>Kurtosis</th>
          <th>Min</th>
          <th>Max</th>
          <th>Median</th>
          <th>25%</th>
          <th>75%</th>
          <th>95%</th>
          <th>99%</th>
          <th>99.9%</th>
        </tr>
      </table>
      <h3>Gauges</h3>
      <table id="gauges">
        <tr id="header"><th>Name</th><th>Value</th><th>Updated</th></tr>
      </table>
      <h3>Meters</h3>
      <table id="meters">
        <tr id="header">
          <th>Name</th><th>Value</th><th>1 min</th>
          <th>5 min</th><th>15 min</th><th>Rates</th>
          <th>1 min</th><th>5 min</th><th>15 min</th>
          <th>Updated</th>
        </tr>
      </table>
    </div>
    <h3>Graphs</h3>
    <div id="graphs"></div>
    <script src="http://d3js.org/d3.v2.js"></script>
    <script src="https://ajax.googleapis.com/ajax/libs/jquery/1.7.2/jquery.min.js"></script>
    <script src="https://raw.github.com/jasondavies/science.js/master/science.v1.min.js"></script>
    <script src="http://square.github.com/cubism/cubism.v1.js"></script>
    <script src="http://masonry.desandro.com/jquery.masonry.min.js"></script>
    <script>

"use strict"

var tableInterval = 1000,
KDEInterval = 10000,
heatmapInterval = 1000

function MetricRow(name){
  this.graphs = {}
  this.graphcount = 0
  this.graphdiv = d3.select("#graphs").append("div").attr("class", "graph")
  this.row = null
  this.name = name
  this.data = null

  this.getDataFunc = function(){
    var m = this
    return function(){ return m.data }
  }

  this.drawRow = function(vals){
    var cells = this.row.selectAll("td").data(vals)
    cells.enter().append("td")
    cells.each(function(d){
      var cell = d3.select(this)
      if($.isFunction(d)){
        d(cell)
      } else {
        if($.isNumeric(d))
          d = formatFixedLen(d)
        cell.text(d)
      }
    })
  }

  this.clickableCell = function(onclick){
    return function(cell){
      cell.on("click", function(){onclick(cell)})
    }
  }

  this.graphToggle = function(graphname, accessor, Constructor, color){
    var m = this
    return this.clickableCell(function(cell){
      if(m.graphs[graphname] != null){
        m.graphcount--
        if(m.graphcount == 0){
          m.graphdiv.style("padding", "0")
          m.graphdiv.select("h3#title").remove()
        }

        m.graphs[graphname].destroy()
        delete m.graphs[graphname]

        cell.style("background-color", "transparent")
      } else {
        if(m.graphcount == 0){
          m.graphdiv.style("padding", "5px")
          m.graphdiv.append("h3").attr("id", "title").text(m.name)
        }
        m.graphcount++

        m.graphs[graphname] = new Constructor(graphname, accessor, m.graphdiv)

        cell.style("background-color", color)
      }
    })
  }

  this.destroy = function(){
    for(var i in this.graphs)
      this.graphs[i].destroy()
    this.graphcount = 0
    this.row.remove()
    this.graphdiv.remove()
  }
}

MetricRowUpdater.metrics = []

function MetricRowUpdater(name, type){
  var row
  this.update = function(){
    $.getJSON("/metric", {"name" : name},
    function(d){
      row.data = d.Value
      row.draw()
    })
  }

  var show
  var hider = d3.select("#hides").append("div")
  var checkbox = hider.append("input").attr("type", "checkbox")
  var t = this
  checkbox.on("click", function(){
    if(show)
      t.hide()
    else
      t.display()
  })

  hider.append("span").text(name)

  this.display = function(){
    show = true
    checkbox.attr("checked", "yes")

    switch(type){
    case "counter":
      row = new CounterRow(name)
      break
    case "distribution":
      row = new DistributionRow(name)
      break
    case "gauge":
      row = new GaugeRow(name)
      break
    case "meter":
      row = new MeterRow(name)
      break
    }
    this.update()
  }

  this.hide = function(){
    show = false
    checkbox.attr("checked", "false")

    row.destroy()
  }

  this.display()

  MetricRowUpdater.metrics.push(this)
}

MetricRowUpdater.updateAll = function(){
  var metrics = MetricRowUpdater.metrics
  for(var i in metrics){
    metrics[i].update()
  }
}

function CounterRow(name){
  MetricRow.apply(this)
  this.name = name
  var data = this.getDataFunc()

  this.draw = function(){
    var updated = new Date(data().LastUpdated)
    var row = [
      this.name,
      data().Value,
      updated.toLocaleTimeString()
    ]

    this.drawRow(row)
  }

  this.makeToggles = function(){
    var row = [
      null,
      this.graphToggle("Value", function(){return data().Value},
        CubismGraph, "yellow"),
      null
    ]

    this.drawRow(row)
  }

  this.row = d3.select("#counters").append("tr")
  this.makeToggles()
}

function DistributionRow(name){
  MetricRow.apply(this)
  this.name = name
  var data = this.getDataFunc()

  var cubismToggles = [
    ["Mean", function(){return data().Mean}],
    ["Variance", function(){return data().Variance}],
    ["StdDev", function(){return data().StandardDeviation}],
    ["Skewness", function(){return data().Skewness}],
    ["Kurtosis", function(){return data().Kurtosis}],
    ["Min", function(){return data().Percentiles[0]}],
    ["Max", function(){return data().Percentiles[7]}],
    ["Median", function(){return data().Percentiles[2]}],
    ["25%", function(){return data().Percentiles[1]}],
    ["75%", function(){return data().Percentiles[3]}],
    ["95%", function(){return data().Percentiles[4]}],
    ["99%", function(){return data().Percentiles[5]}],
    ["99.9%", function(){return data().Percentiles[6]}],
  ]

  this.draw = function(){
    var name = this.name
    var row = [name, data().Count]
    for(var i = 0; i < cubismToggles.length; i++){
      row.push(cubismToggles[i][1]())
    }

    this.drawRow(row)
  }

  this.makeToggles = function(){
    var name = this.name
    var histogramAccessor = function(){
      return {
          name: name,
          window: data().Window,
          range: data().RangeHint
      }
    }

    var row = [
      this.graphToggle("heatmap", histogramAccessor,
        HeatmapUpdater, "lightgreen"),
      this.graphToggle("samples", histogramAccessor,
        KDEGraphUpdater, "lightgreen")
    ]
    for(var i = 0; i < cubismToggles.length; i++){
      row.push(this.graphToggle(cubismToggles[i][0],
        cubismToggles[i][1], CubismGraph, "lightgreen"))
    }

    this.drawRow(row)
  }

  this.row = d3.select("#distributions").append("tr")
  this.makeToggles()
}

function GaugeRow(name){
  MetricRow.apply(this)
  this.name = name
  var data = this.getDataFunc()

  this.draw = function(){
    var updated = new Date(data().LastUpdated)
    var str
    if(data().Value.length > 80) {
      str = slice(data().Value, 77) + "..."
    } else {
      str = data().Value
    }
    this.drawRow([name, str, updated.toLocaleTimeString()])
  }

  this.row = d3.select("#gauges").append("tr")
}

function MeterRow(name){
  MetricRow.apply(this)
  this.name = name
  var data = this.getDataFunc()

  var toggles = [
    ["Value", function(){return data().Value}],
    ["1 min avg", function(){return data().Derivatives[0][1]}],
    ["5 min avg", function(){return data().Derivatives[0][2]}],
    ["15 min avg", function(){return data().Derivatives[0][3]}],
    ["instant rate", function(){return data().Derivatives[1][0]}],
    ["1 min rate", function(){return data().Derivatives[1][1]}],
    ["5 min rate", function(){return data().Derivatives[1][2]}],
    ["15 min rate", function(){return data().Derivatives[1][3]}]
  ]

  this.draw = function(){
    var row = [this.name]
    for(var i = 0; i < toggles.length; i++){
      row.push(toggles[i][1]())
    }
    var updated = new Date(data().LastUpdated)
    row.push(updated.toLocaleTimeString())

    this.drawRow(row)
  }

  this.makeToggles = function(){
    var row = [null]
    for(var i = 0; i < toggles.length; i++){
      row.push(this.graphToggle(toggles[i][0], toggles[i][1],
        CubismGraph, "pink"))
    }
    row.push(null)

    this.drawRow(row)
  }

  this.row = d3.select("#meters").append("tr")
  this.makeToggles()
}

HeatmapUpdater.heatmaps = []

function HeatmapUpdater(name, accessor, div){
  var n, interval, popPerRow, heatmap

  this.preload = function(arr){
    var u = this
    $.getJSON("/metric",
      {"name" : accessor().name,
       "samples" : true,
       "limit" : 1000},
      function(d){
        d = d.Value
        var range = accessor().range
        if(range[0] == range[1]) {
          var range = dataRange(d.Samples, 0.2, 0.8, 2)[0]
        }

        heatmap = new Heatmap(div, range)
        n = heatmap.getHRes()

        var w = accessor().window/1000000
        interval = Math.floor(w/n/heatmapInterval)
        if(interval == 0)
          interval++
        interval *= heatmapInterval

        var now = new Date
        for(var i = n; i >= 0; i--){
          if(interval * i > w)
            continue
          u.update(new Date(now.getTime() - interval * i))
        }

        HeatmapUpdater.heatmaps.push([u, interval, 0])
    })
  }

  this.update = function(end){
    var begin = new Date(end.getTime() - interval)

    var newest = end
    var oldest = new Date(newest.getTime() - (n- 1) * interval)

    $.getJSON("/metric",
      {"name" : accessor().name,
       "samples" : true,
       "limit" : 100,
       "begin" : begin.toISOString(),
       "end" : end.toISOString()},
      function(d){
        if(d.Value.Count < 0)
          return

        heatmap.draw(d.Value.Samples, newest, oldest, d.Value.Count)
      })
  }

  this.preload()

  this.destroy = function(data){
    var h = HeatmapUpdater.heatmaps
    for(var i in h){
      if(h[i][0] == this){
        h.splice(i, 1)
        break
      }
    }
    heatmap.destroy()
  }
}

HeatmapUpdater.updateAll = function(){
  var h = HeatmapUpdater.heatmaps
  for(var i in h){
    h[i][2] -= heatmapInterval
    if(h[i][2] <= 0){
      h[i][0].update(new Date)
      h[i][2] = h[i][1]
    }
  }
}

function Heatmap(div, range){
  var w = 345, wmargin = 60, h = 100, hmargin = 20
  var shiftw = 5, nbins = 20
  var newtime, oldtime
  var hists = []

  var owndiv = div.append("div")
  var canvas = owndiv.append("canvas").attr("id", "heatmap")
    .attr("width", w + wmargin).attr("height", h + hmargin)
  var ctx = canvas.node().getContext('2d')

  ctx.fillStyle = "black"
  ctx.fillRect(0, 0, w, h)

  var mousemovediv = owndiv.append("div").attr("id", "mousevalue")
  var mouse

  var realSizeTotal = 0

  var cleanMousemove = function(){
    mouse = null
    mousemovediv.text("")
  }

  var updateMousemove = function(){
    if(mouse == null)
      return
    var x = mouse[0]
    var y = mouse[1]

    if(x >= w || y >= h){
      cleanMousemove()
      return
    }

    var i = Math.floor(x/shiftw) + hists.length - w/shiftw

    if(i < 0){
      str = "No data"
      mousemovediv.text(str)
      return
    }

    var t = new Date(x/w * (newtime - oldtime) + oldtime)
    var j = Math.floor((h - y - 1)/h * nbins)
    var str = "T: " + t.toLocaleTimeString() + " "

    var bin = hists[i][j]
    str += "Rng: [" + Math.floor(bin.x).toPrecision(3) + ", "
    if(j < nbins - 1)
      str += Math.floor(hists[i][j+1].x).toPrecision(3) + ") "
    else
      str += Number.POSITIVE_INFINITY + ") "

    str += "N: " + bin.y.toPrecision(3)
    str += " Bin: " + Math.floor(100 * bin.y / hists[i].total) + "%"
    str += " Tot: " + (100 * bin.y / realSizeTotal).toPrecision(3) + "%"

    mousemovediv.text(str)
  }

  canvas.on("mousemove", function(){
    mouse = d3.mouse(this)
    updateMousemove(mouse)
  })

  canvas.on("mouseout", cleanMousemove)

  var colormap = function(v){
    var lightness = 15 + 60 * v
    var hue = 240 - 360 * v
    return "hsl(" + hue + ", 100%, " + lightness + "%)"
  }

  var thresholds = [Number.NEGATIVE_INFINITY]
  for(var i = 0; i < nbins - 1; i++)
    thresholds.push((range[1] - range[0]) * i / (nbins - 2) + range[0])
  thresholds.push(Number.POSITIVE_INFINITY)

  var hist = d3.layout.histogram().bins(thresholds)

  this.draw = function(arr, newest, oldest, realsize){
    var bins = hist(arr)
    bins.total = realsize

    if(arr.length != 0){
      for(i in bins){
        bins[i].y *= realsize/arr.length
      }
    }

    hists.push(bins)
    realSizeTotal += realsize
    while(hists.length > w/shiftw){
      var v = hists.shift()
      realSizeTotal -= v.total
    }

    ctx.drawImage(ctx.canvas, shiftw, 0, w - shiftw, h, 0, 0, w - shiftw, h)

    ctx.save()
    ctx.translate(w - shiftw, 0)
    for(var i = 0; i < nbins; i++){
      if(realSizeTotal == 0)
        ctx.fillStyle = colormap(0)
      else
        ctx.fillStyle = colormap(bins[i].y / realSizeTotal * hists.length)

      ctx.fillRect(0, (nbins - i - 1) * h / nbins, shiftw, h / nbins)
    }
    ctx.fillStyle = "red"
    ctx.fillRect(0, h / nbins - 1, shiftw, 1)
    ctx.fillRect(0, (nbins - 1) * h / nbins, shiftw, 1)
    ctx.restore()

    ctx.save()
    ctx.fillStyle = "black"
    ctx.font = "12px sans-serif"

    // x axis
    ctx.save()
    ctx.translate(0, h)
    ctx.clearRect(0, 0, w, hmargin)
    newtime = newest.getTime()
    oldtime = oldest.getTime()
    var mid = new Date((newtime + oldtime) / 2)
    ctx.fillText(oldest.toLocaleTimeString(), 0, 15)
    ctx.fillText(mid.toLocaleTimeString(), (w-50)/2, 15)
    ctx.fillText(newest.toLocaleTimeString(), w-50, 15)
    ctx.restore()

    // y axis
    ctx.save()
    ctx.translate(w, 0)
    ctx.clearRect(0, 0, wmargin, h)
    ctx.fillText(formatFixedLen(range[1]), 5, 20)
    ctx.fillText(formatFixedLen((range[0] + range[1])/2),
      5, Math.floor((h + 10)/2))
    ctx.fillText(formatFixedLen(range[0]), 5, h - 10)
    ctx.restore()
    ctx.restore()

    updateMousemove(mouse)
  }

  this.getHRes = function(){
    return w/shiftw
  }

  $("#graphs").masonry("reload")
  this.destroy = function(data){
    canvas.remove()
    $("#graphs").masonry("reload")
  }
}

var cubismContext = cubism.context()
    .serverDelay(1000).step(tableInterval).size(400)

function CubismGraph(name, accessor, div){
  var context = cubismContext

  // Force Cubism to synchronously update
  // Cubism reads with start incrementing by step,
  // and stop = start + 6*step. So, it reads everything
  // 7 times. This is not good since the previous values
  // aren't stored on the server.
  var queue = []
  var metric = context.metric(
    function(start, stop, step, callback){
      start = +start
      stop = +stop

      var res = []

      queue.push(accessor())
      if(queue.length > 7)
        queue.shift()

      for(var i = 0; start < stop; i++){
        start += step
        res.push(queue[i])
      }
      callback(null, res)
    }, name)


  var graph = div.append("div").attr("id", name)

  var axis = graph.append("div").attr("class", "axis")
    .call(context.axis().ticks(d3.time.minutes, 2).orient("top"))

  var horizon = graph.selectAll(".horizon")
    .data([metric]).enter()
    .append("div")
    .attr("class", "horizon")
    .call(context.horizon())

  var rule = graph.append("div").data([]).attr("class", "rule")
    .call(context.rule())

  $("#graphs").masonry("reload")

  this.destroy = function(){
    graph.select(".axis")
      .call(context.axis().remove).remove()
    graph.select(".horizon")
      .call(context.horizon().remove).remove()
    graph.select(".rule")
      .call(context.rule().remove).remove()
    graph.remove()
    $("#graphs").masonry("reload")
  }
}

function KDEGraphUpdater(name, accessor, div){
  var graph = new KDEGraph(div)

  this.update = function(){
    $.getJSON("/metric",
      {"name" : accessor().name, "samples" : true, "limit" : 1000},
      function(d){
        graph.draw(d.Value.Samples, accessor().range)
      })
  }
  this.update()

  KDEGraphUpdater.graphs.push(this)

  this.destroy = function(){
    graph.destroy()
    var g = KDEGraphUpdater.graphs
    for(var i in g){
      if(g[i] == this){
        g.splice(i, 1)
        break
      }
    }
  }
}

KDEGraphUpdater.graphs = []

function KDEGraph(div){
  var w = 345, h = 240, wmargin = 55, hmargin = 20

  var svg = div.append("svg").attr("id", "samples")
    .attr("width", w + wmargin)
    .attr("height", h + hmargin)

  var path = svg.append("path").style("fill", "steelblue")
  var xgroup = svg.append("g").attr("id", "xlabel")

  var nbins = 20
  var kdearr, hist, yscl, xscl

  var mousemovediv = div.append("div").attr("id", "mousevalue")
  var mousemoveline = svg.append("line").attr("id", "mouseline")
    .attr("y1", 0).attr("y2", h)
    .style("stroke", "red").style("display", "none")
  var mouse

  var cleanMousemove = function(){
    mouse = null
    mousemovediv.text(null)

    mousemoveline.style("display", "none")
  }

  var updateMousemove = function(){
    if(mouse == null)
      return
    var x = mouse[0]
    var y = mouse[1]

    if(x >= w || y >= h){
      cleanMousemove()
      return
    }

    mousemoveline.style("display", "inline")
    var v = kdearr[x]
    var str = "V: " + formatFixedLen(v[0])
    str += " P: " + formatFixedLen(v[1])
    mousemovediv.text(str)
    mousemoveline.attr("x1", mouse[0]).attr("x2", mouse[0])
  }

  svg.on("mousemove", function(){
    mouse = d3.mouse(this)
    updateMousemove()
  })

  svg.on("mouseout", cleanMousemove)

  this.draw = function(data, range){

    if(range[0] == range[1]){
      var rangeslice = dataRange(data, 0.1, 0.9, 2)
      var range = rangeslice[0]
      var ind = rangeslice[1]
      data = data.slice(ind[0], ind[1])
    } else {
      data = data.sort(function(a,b){return a-b})
      var l = data.length
      for(var i = 0; i < l; i++){
        if(data[i] < range[0])
          data.shift()
      }
      for(var i = data.length - 1; i > 0; i--){
        if(data[i] > range[1])
          data.pop()
      }
    }

    var diff = range[1] - range[0]

    var kde = science.stats.kde().sample(data)
    kde.bandwidth(science.stats.bandwidth.nrd0)
    kdearr = kde(d3.range(range[0], range[1] + diff/2/w, diff/w))

    var ymax = d3.max(kdearr, function(d){ return d[1] })
    xscl = d3.scale.linear().domain(range).range([0, w])
    yscl = d3.scale.linear().domain([0, ymax]).nice().range([h, 0])
      .clamp(true)

    var line = d3.svg.line()
      .x(function(d) { return xscl(d[0]) })
      .y(function(d) { return yscl(d[1]) })

    path.transition().duration(1000).attr("d", function(d){
      kdearr.push([range[1], 0])
      kdearr.unshift([range[0], 0])
      return line(kdearr)
    })

    var ticks = yscl.ticks(5)
    var rules = svg.selectAll("g.rule").data(ticks)
    var newrules = rules.enter().append("g").attr("class", "rule")
    newrules.append("line").attr("x2", w).attr("stroke", "gray")
    newrules.append("text")
    rules.exit().remove()
    rules.attr("transform", function(d){return "translate(0," + yscl(d) + ")"})

    var ylabels = rules.selectAll("text")
    ylabels.attr("x", w + 5).attr("dy", ".35em")
      .text(formatFixedLen)

    var xlabels = xgroup.selectAll("text")
      .data([range[0], (range[0]+range[1])/2, range[1]])
    xlabels.enter().append("text")
      .attr("x", function(d, i){ return i * (w - 50)/2 })
      .attr("y", h + 10)
      .attr("dy", ".35em")
    xlabels.text(formatFixedLen)

    updateMousemove()
  }

  $("#graphs").masonry("reload")
  this.destroy = function(){
    svg.remove()
    $("#graphs").masonry("reload")
  }
}

KDEGraphUpdater.updateAll = function(){
  var g = KDEGraphUpdater.graphs
  for(var i in g){
    g[i].update()
  }
}



function identity(d){
  return d
}

function formatDuration(t){
  if(t == 0)
    return "0"

  units = ["ns", "us", "ms"]
  for(var i = 0; i < 3; i++){
    if(t < 1000){
      return t.toString() + units[i]
    }
    t /= 1000
  }
  if(t < 60){
    return t.toString() + "s"
  }
  t /= 60
  return t.toString() + "m"
}

function formatFixedLen(n){
  if(isNaN(n))
    return n

  var fix = n.toPrecision(6)
  if(fix.length <= 7){
    return fix
  } else {
    return n.toExponential(2)
  }
}


function dataRange(data, pmin, pmax, mul){
  data = data.sort(function(a,b){return a-b})
  var imin = Math.floor(data.length*pmin)
  var imax = Math.ceil(data.length*pmax)
  var smin = data[imin]
  var smax = data[imax - 1]
  if(smin == smax){
    if(smax < 0){
      smin = 2 * smin
      smax = 0
    } else if(smax == 0){
      smin = -10
      smax = 10
    } else {
      smin = 0
      smax = 2 * smax
    }
  }

  var adj = (smax - smin) * (mul - 1)/2
  smax += adj
  smin -= adj

  return [[smin, smax], [imin, imax]]
}

function fixGraphLayout(){
  $("#graphs").masonry("reload")
}

$(document).ready(function(){
  var showhides = false

  var togglehides = function(){
    var t = d3.select(this)
    if(showhides){
      d3.select("#hides").style("display", "none")
      t.text("Show metric show/hide list")
    } else {
      d3.select("#hides").style("display", "inline")
      t.text("Hide metric show/hide list")
    }
    showhides = !showhides
  }

  d3.select("#hidehides").on("click", togglehides)
    .each(togglehides)

  $.getJSON("/list", {}, function(list){
    list.sort()
    var throwaway = []
    for(var i in list){
      throwaway.push(new MetricRowUpdater(list[i][0], list[i][1]))
    }

    $("#graphs").masonry({
      itemSelector: ".graph",
      columnWidth: 410
    })

    setInterval(function(){MetricRowUpdater.updateAll()}, tableInterval)
    setInterval(function(){KDEGraphUpdater.updateAll()}, KDEInterval)
    setInterval(function(){HeatmapUpdater.updateAll()}, heatmapInterval)
  })

})

    </script>
  </body>
`