"use strict"

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
    ctx.clearRect(0, 0, wmargin, h + hmargin)
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
