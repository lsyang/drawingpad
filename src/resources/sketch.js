ID=-1
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
        // $("#output").append("Waiting for system time..");
         setInterval(function() {sketch_object.getUpdate()}, 200);
       });

  Sketch = (function() {
    function Sketch(el, opts) {
      this.register()
      this.el = el;
      this.canvas = $(el);
      this.context = el.getContext('2d');
      this.options = $.extend({
        toolLinks: true,
        defaultTool: 'marker',
        defaultColor: '#000000',
        defaultSize: 3
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

  Sketch.prototype.register=function(){
    $.post(window.location.origin+"/register", {id: ID}, function(data, status) {
       ID=data //this also hear the response to other clients
    });
  };

  Sketch.prototype.getUpdate=function() {
    console.log("get updating")
    if (ID==-1){
      this.register()
    }
   $.post(window.location.origin+"/drawUpdate", {id: ID}, function(data, status) {
    //status = success

   obj = JSON.parse(data);
   //console.log(obj)
   var sketch = sketch_object;
   if(obj.Has_operation){
    console.log("getting update")
    for (var i=0; i<obj.New_operations.length; i++){
      var op=obj.New_operations[i]
      //if op.OpName=="Put"{
      //console.log(op)
      sketch.executeDraw(op.ClientStroke.Start_x,op.ClientStroke.Start_y,op.ClientStroke.End_x,op.ClientStroke.End_y,op.ClientStroke.Color,op.ClientStroke.Size)
    //}
    }
   }
  });
  };

  Sketch.prototype.executeDraw=function(x1,y1,x2,y2,col,size){
    var sketch = sketch_object;
    sketch.el.width = sketch.canvas.width();
    sketch.context = sketch.el.getContext('2d');
      var action = {
        tool: 'marker',
        color: col, 
        size: size,
        events: []
      };
      action.events.push({
        x: x1,
        y: y1,
      });
      action.events.push({
        x: x2,
        y: y2,
      });
      sketch.actions.push(action)
      //console.log("drawing")
      sketch.redraw()
    };


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
          console.log("redrawing")
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
         var val=this.color;
         var brush_size=this.size;
       //   //testing purpose
       //  // var keys=JSON.stringify([this.xy2val(_x,_y)])
      // console.log(Math.round(pre.x),Math.round(pre.y),Math.round(_x),Math.round(_y));
        $.post(window.location.origin+"/stroke", {id: ID, startx: Math.round(pre.x),starty: Math.round(pre.y), endx: Math.round(_x),endy: Math.round(_y), color: val, size: brush_size}, function(data, status) {
             
        });
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