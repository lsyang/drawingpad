var __slice = Array.prototype.slice;
(function($) {
  var Sketch;
  $.fn.sketch = function() {
    var args, key, sketch;
    key = arguments[0], args = 2 <= arguments.length ? __slice.call(arguments, 1) : [];
    if (this.length > 1) {
      $.error('Sketch.js can only be called on one element at a time.');
    }
    sketch = this.data('sketch');
    if (typeof key === 'string' && sketch) {
      if (sketch[key]) {
        if (typeof sketch[key] === 'function') {
          return sketch[key].apply(sketch, args);
        } else if (args.length === 0) {
          return sketch[key];
        } else if (args.length === 1) {
          return sketch[key] = args[0];
        }
      } else {
        return $.error('Sketch.js did not recognize the given command.');
      }
    } else if (sketch) {
      return sketch;
    } else {
      this.data('sketch', new Sketch(this.get(0), key));
      return this;
    }
  };

          //call post every 1s
      $(document).ready(function () {
         $("#output").append("Waiting for system time..");
         //setInterval(function() {sketch_object.getUpdate()}, 1000);
       });

  Sketch = (function() {
    function Sketch(el, opts) {
      this.el = el;
      this.canvas = $(el);
      this.context = el.getContext('2d');
      this.options = $.extend({
        toolLinks: true,
        defaultTool: 'marker',
        defaultColor: '#000000',
        defaultSize: 1
      }, opts);
      this.painting = false;
      this.color = this.options.defaultColor;
      this.size = this.options.defaultSize;
      this.tool = this.options.defaultTool;
      this.actions = [];
      this.action = [];
      this.history=[];
      this.canvas.bind('click mousedown mouseup mousemove mouseleave mouseout touchstart touchmove touchend touchcancel', this.onEvent);
      sketch_object=this
      if (this.options.toolLinks) {
        $('body').delegate("a[href=\"#" + (this.canvas.attr('id')) + "\"]", 'click', function(e) {
          var $canvas, $this, key, sketch, _i, _len, _ref;
          $this = $(this);
          $canvas = $($this.attr('href'));
          sketch = $canvas.data('sketch');
          _ref = ['color', 'size', 'tool'];
          for (_i = 0, _len = _ref.length; _i < _len; _i++) {
            key = _ref[_i];
            if ($this.attr("data-" + key)) {
              sketch.set(key, $(this).attr("data-" + key));
            }
          }
          if ($(this).attr('data-download')) {
            sketch.download($(this).attr('data-download'));
          }
          return false;
        });
      }
    }


  Sketch.prototype.getUpdate=function() {
  console.log("hahaha")

   $.post(window.location.origin+"/drawUpdate", "", function(data, status) {
    //status = success
   obj = JSON.parse(data);
   //console.log(obj)
   var sketch = sketch_object;
   if(obj.Has_map){ //clear canvas
    sketch.clear()
    //draw canvas
    for(var i=0; i<obj.Board.length; i++){
      sketch.executeDraw(i,obj.Board[i])
    }
   }
   if(obj.Has_operation){
    for (var i=0; i<obj.New_operations.length; i++){
      var op=obj.New_operations[i]
      sketch.executeDraw(op.Key,op.Val)
    }
   }
  });
  };

  Sketch.prototype.clear=function(){
    var self = this;
    this.actions = [];
    self.context.clearRect(0, 0, self.canvas.width(), self.canvas.height());
  }
  Sketch.prototype.xy2val=function(x,y){
    var self=this;
    x=Math.round(x)
    y=Math.round(y)
    return Math.round(self.canvas.width())*y+x
  }
  Sketch.prototype.val2xy=function(val){
    var self=this;
    var y=Math.floor(val/Math.round(self.canvas.width()))
    var x=val-y*Math.round(self.canvas.width())
    return [x, y]
  }
  Sketch.prototype.listPoints=function(x1,y1,x2,y2){
    var c=0
    var result=[]
    var deltax=1
    var deltay=1
    var val=this.color;
    if(y1>y2){
      deltay=-1
    }
    if(x1>x2){
      deltax=-1
    }
    if (x1==x2){
      for(var i=0; i<=(y2-y1)*deltay; i++){
        this.put(this.xy2val(x1,y1+i*deltay),val)
        //result.push(this.xy2val(x1,y1+i*deltay))
      }
    }else{
      c=(y2-y1)/(x2-x1);
      if (Math.abs(c)<1){ //when its almost straight line
        for(var i=0; i<=(x2-x1)*deltax; i++){
          this.put(this.xy2val(x1+deltax*i,y1+c*deltax*i), val)
          //result.push(this.xy2val(x1+deltax*i,y1+c*deltax*i))
        }
      }else{
        for(var i=0; i<=(y2-y1)*deltay; i++){
          this.put(this.xy2val(x1+1/c*deltay*i,y1+deltay*i),val)
          //result.push(this.xy2val(x1+1/c*deltay*i,y1+deltay*i))
        }
      }
    }
    return result
  }
  Sketch.prototype.put=function(key, val){
    //put key value into the server
    $.post(window.location.origin+"/stroke", {Key: key, Value: val}, function(data, status) {
         obj = JSON.parse(data);
         //console.log(obj)
         var sketch = sketch_object;
         if(obj.Has_map){ //clear canvas
          sketch.clear()
          //draw canvas
          for(var i=0; i<obj.Board.length; i++){
            //sketch.executeDraw(i,obj.Board[i])
          }
         }
         if(obj.Has_operation){
          for (var i=0; i<obj.New_operations.length; i++){
            var op=obj.New_operations[i]
            sketch.executeDraw(op.Key,op.Value)
          }
         }
    });
  }
  Sketch.prototype.executeDraw=function(pos, col){
    var sketch = sketch_object;
    sketch.el.width = sketch.canvas.width();
    sketch.context = sketch.el.getContext('2d');
      var action = {
        tool: 'marker',
        color: col, 
        size: 1,
        events: []
      };
      var key=sketch.val2xy(pos)
      var _x=key[0]
      var _y=key[1]
      action.events.push({
        x: _x,
        y: _y,
      });
      action.events.push({
        x: _x,
        y: _y+1,
      });
      sketch.actions.push(action)
      sketch.redraw()
    }
    Sketch.prototype.test=function(x1,y1,x2,y2){

        var lala=this.listPoints(x1,y1,x2,y2)

        for (var i=0; i<lala.length; i++){
          // console.log(lala[i])
           //console.log(this.val2xy(lala[i]));
          this.executeDraw(lala[i],'#000000') 
        }
    }


    Sketch.prototype.download = function(format) {
      var mime;
      format || (format = "png");
      if (format === "jpg") {
        format = "jpeg";
      }
      mime = "image/" + format;
      return window.open(this.el.toDataURL(mime));
    };
    Sketch.prototype.set = function(key, value) {
      this[key] = value;
      return this.canvas.trigger("sketch.change" + key, value);
    };
    Sketch.prototype.startPainting = function() {
      this.painting = true;
      // return this.action = {
      //   tool: this.tool,
      //   color: this.color,
      //   size: parseFloat(this.size),
      //   events: []
      // };
    };
    Sketch.prototype.stopPainting = function() {
      // if (this.action) {
      //   this.actions.push(this.action);
      // }
      this.painting = false;
      //this.action = null;
      this.history=[];
      return this.redraw();
    };
    Sketch.prototype.onEvent = function(e) {
      if (e.originalEvent && e.originalEvent.targetTouches) {
        e.pageX = e.originalEvent.targetTouches[0].pageX;
        e.pageY = e.originalEvent.targetTouches[0].pageY;
      }
      $.sketch.tools[$(this).data('sketch').tool].onEvent.call($(this).data('sketch'), e);
      e.preventDefault();
      return false;
    };

    Sketch.prototype.redraw = function() {
      //console.log("redraw")
      var sketch;
      this.el.width = this.canvas.width();
      this.context = this.el.getContext('2d');
      sketch = this;
      $.each(this.actions, function() {
        if (this.tool) {
          return $.sketch.tools[this.tool].draw.call(sketch, this);
        }
      });
      // if (this.painting && this.action) {
      //   return $.sketch.tools[this.action.tool].draw.call(sketch, this.action);
      // }
    };
    return Sketch;
  })();

  $.sketch = {
    tools: {}
  };


  $.sketch.tools.marker = {

    onEvent: function(e) {
      switch (e.type) {
        case 'mousedown':
        case 'touchstart':
          this.startPainting();
          break;
        case 'mouseup':
        case 'mouseout':
        case 'mouseleave':
        case 'touchend':
        case 'touchcancel':
          this.stopPainting();
      }
      if (this.painting) {
        
        var _x=e.pageX - this.canvas.offset().left;
        var _y=e.pageY - this.canvas.offset().top;
        var action=this.history
        action.push({
          x: _x,
          y: _y,
        })
        //define pre as the previous point
        if (action.length>1){
          var pre=action[action.length-2]
        }else{
          var pre=action[action.length-1] //current point
        }
        var key=this.listPoints(Math.round(pre.x),Math.round(pre.y),Math.round(_x),Math.round(_y)) //put in integers
  //orginal stuff
       // var keys=JSON.stringify(key)
       //   var val=this.color;
       //   //testing purpose
       //  // var keys=JSON.stringify([this.xy2val(_x,_y)])
       //  $.post(window.location.origin+"/stroke", {Key: keys, Value: val}, function(data, status) {
       //       obj = JSON.parse(data);
       //       //console.log(obj)
       //       var sketch = sketch_object;
       //       if(obj.Has_map){ //clear canvas
       //        sketch.clear()
       //        //draw canvas
       //        for(var i=0; i<obj.Board.length; i++){
       //          //sketch.executeDraw(i,obj.Board[i])
       //        }
       //       }
       //       if(obj.Has_operation){
       //        for (var i=0; i<obj.New_operations.length; i++){
       //          var op=obj.New_operations[i]
       //          sketch.executeDraw(op.Key,op.Value)
       //        }
       //       }
       //  });
        //orginal code
        // this.action.events.push({
        //   x: e.pageX - this.canvas.offset().left,
        //   y: e.pageY - this.canvas.offset().top,
        //   event: e.type
        // });
        // return this.redraw();
      }
    },
    draw: function(action) {
      var event, previous, _i, _len, _ref;
      this.context.lineJoin = "round";
      this.context.lineCap = "round";
      this.context.beginPath();
      this.context.moveTo(action.events[0].x, action.events[0].y);
      

      _ref = action.events;
      for (_i = 0, _len = _ref.length; _i < _len; _i++) {
        event = _ref[_i];
        this.context.lineTo(event.x, event.y);
        //console.log(event.x, event.y)
        previous = event;
      }
      this.context.strokeStyle = action.color;
      this.context.lineWidth = action.size;      
      // console.log(event.x, event.y, action.color)
      return this.context.stroke();
    }
  };
  return $.sketch.tools.eraser = {
    onEvent: function(e) {
      return $.sketch.tools.marker.onEvent.call(this, e);
    },
    draw: function(action) {
      var oldcomposite;
      oldcomposite = this.context.globalCompositeOperation;
      this.context.globalCompositeOperation = "copy";
      action.color = "rgba(0,0,0,0)";
      $.sketch.tools.marker.draw.call(this, action);
      return this.context.globalCompositeOperation = oldcomposite;
    }
  };
})(jQuery);