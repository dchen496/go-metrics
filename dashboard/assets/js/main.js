"use strict"

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
    MetricRowUpdater.updateAll()

    $("#graphs").masonry({
      itemSelector: ".graph",
      columnWidth: 410
    })

    setInterval(function(){MetricRowUpdater.updateAll()}, tableInterval)
    setInterval(function(){KDEGraphUpdater.updateAll()}, KDEInterval)
    setInterval(function(){HeatmapUpdater.updateAll()}, heatmapInterval)
  })

})


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
        cell.html(d)
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
  this.update = function(allmap){
    row.data = allmap[name].Value
    row.draw()
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
  $.getJSON("/all", null, function(d){
    for(var i in metrics){
      metrics[i].update(d)
    }
  })
}

function CounterRow(name){
  MetricRow.apply(this)
  this.name = name
  var data = this.getDataFunc()

  this.draw = function(){
    var row = [
      this.name,
      data().Value,
    ]

    this.drawRow(row)
  }

  this.makeToggles = function(){
    var row = [
      null,
      this.graphToggle("Value", function(){return data().Value},
        CubismGraph, "yellow"),
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
    //wrap text
    var str = data().Value
    var wrapped = ""
    for(var i = 0; i < str.length; i += 100){
      wrapped += str.slice(i, i+100) + "<br />"
    }
    this.drawRow([name, wrapped])
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
