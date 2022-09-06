function Sequence(startValues, steps) {
    this.startValues = startValues;
    this.steps = steps;
    
    this.previousStep = -1;
    this.totalDuration = steps.reduce(function(acc, step) { return acc + step.duration; }, 0);
}

Sequence.prototype.do = function(time) {
    var stepIndex = 0;
    for (; stepIndex < this.steps.length; stepIndex++) {
        if (time <= this.steps[stepIndex].duration) {
            break;
        }
        
        time -= this.steps[stepIndex].duration;
    }

    if (stepIndex >= this.steps.length) {
        return;
    }

    var step = this.steps[stepIndex];
    var st = time / step.duration;
    if (step.easingFunction) {
        st = step.easingFunction(st);
    }

    var vals = step.values || {};
    var newVals = {};
    Object.keys(vals).forEach((function(key) {
        var after = vals[key];
        var before = vals[key];
        if (stepIndex === 0) {
            if (this.startValues.hasOwnProperty(key)) {
                before = this.startValues[key];
            }
        } else {
            var prevStep = this.steps[stepIndex - 1];
            if (prevStep.values && prevStep.values.hasOwnProperty(key)) {
                before = prevStep.values[key];
            }
        }

        var newVal = after;
        if (typeof after === 'number') {
            newVal = before * (1 - st) + after * st;
        } else if (after.clone && after.lerp) {
            newVal = before.clone().lerp(after, st);
        }

        newVals[key] = newVal;
    }).bind(this));

    if (step.init && stepIndex !== this.previousStep) {
        step.init();
    }
    step.f(newVals, st);

    this.previousStep = stepIndex;
}

function transformVerts(obj, transformFunc) {
    if (obj.geometry && obj.geometry.originalVertices) {
        for (var i = 0; i < obj.geometry.vertices.length; i++) {
            var orig = obj.geometry.originalVertices[i];
            obj.geometry.vertices[i] = transformFunc(orig);
        }
        obj.geometry.verticesNeedUpdate = true;
    }

    obj.children.forEach(function(child) {
        transformVerts(child, transformFunc);
    });
}
function resetVerts(obj) {
    transformVerts(obj, function(orig) {
        return orig;
    });
}

function transformPoints(obj, transformFunc) {
    if (obj.isTgPoint) {
        transformFunc(obj);
    }

    obj.children.forEach(function(child) {
        transformPoints(child, transformFunc);
    });
}
function resetPoints(obj) {
    transformPoints(obj, function(point) {
        point.position.x = point.originalPosition.x;
        point.position.y = point.originalPosition.y;
        point.position.z = point.originalPosition.z;
    });
}

function setOpacity(obj, opacity) {
    if (obj.material) {
        obj.material.opacity = opacity;
    }

    obj.children.forEach(function(child) {
        setOpacity(child, opacity);
    });
}

function wobble(t, magnitude, speed) {
    return toRadians(magnitude) * Math.sin(speed * t);
}

/*
 * From https://gist.github.com/gre/1650294
 * Easing Functions - inspired from http://gizma.com/easing/
 * only considering the t value for the range [0, 1] => [0, 1]
 */
Easing = {
  // no easing, no acceleration
  linear: function (t) { return t },
  // accelerating from zero velocity
  easeInQuad: function (t) { return t*t },
  // decelerating to zero velocity
  easeOutQuad: function (t) { return t*(2-t) },
  // acceleration until halfway, then deceleration
  easeInOutQuad: function (t) { return t<.5 ? 2*t*t : -1+(4-2*t)*t },
  // accelerating from zero velocity 
  easeInCubic: function (t) { return t*t*t },
  // decelerating to zero velocity 
  easeOutCubic: function (t) { return (--t)*t*t+1 },
  // acceleration until halfway, then deceleration 
  easeInOutCubic: function (t) { return t<.5 ? 4*t*t*t : (t-1)*(2*t-2)*(2*t-2)+1 },
  // accelerating from zero velocity 
  easeInQuart: function (t) { return t*t*t*t },
  // decelerating to zero velocity 
  easeOutQuart: function (t) { return 1-(--t)*t*t*t },
  // acceleration until halfway, then deceleration
  easeInOutQuart: function (t) { return t<.5 ? 8*t*t*t*t : 1-8*(--t)*t*t*t },
  // accelerating from zero velocity
  easeInQuint: function (t) { return t*t*t*t*t },
  // decelerating to zero velocity
  easeOutQuint: function (t) { return 1+(--t)*t*t*t*t },
  // acceleration until halfway, then deceleration 
  easeInOutQuint: function (t) { return t<.5 ? 16*t*t*t*t*t : 1+16*(--t)*t*t*t*t }
}
