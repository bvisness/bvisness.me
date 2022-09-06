function toRadians(angle) {
    return angle * (Math.PI / 180);
}

var ThreeGraph = {
    Grid: function Grid() {
        return new THREE.GridHelper(10, 10);
    },
    Axes: function Axes(size = 3) {
        var floor = new THREE.GridHelper(size * 2, size * 2);

        var yAxis = ThreeGraph.Line([
            new THREE.Vector3(0, -size, 0),
            new THREE.Vector3(0, size, 0),
        ]);
        floor.add(yAxis);

        return floor;
    },
    Line: function Line(points, material = ThreeGraph.Mat_Color()) {
        var geometry = new THREE.Geometry();
        geometry.originalVertices = [];
        points.forEach(function (point) {
            geometry.vertices.push(point);
            geometry.originalVertices.push(point);
        });

        var line = new THREE.Line(geometry, material);
        line.computeLineDistances();

        return line;
    },
    PointSlopeLine: function PointSlopeLine(point, direction, material = ThreeGraph.Mat_Color()) {
        var bigDir = direction.clone().multiplyScalar(20);
        return ThreeGraph.Line([
            point.clone().sub(bigDir),
            point.clone().add(bigDir),
        ], material);
    },
    Camera: function Camera(fov, aspect) {
        var w = 0.5;
        var d = 0.5 / Math.tan(toRadians(fov));
        var h = 0.5 / aspect;
        var mainBody = ThreeGraph.Line([
            new THREE.Vector3(0, 0, 0),
            new THREE.Vector3(w, h, d),
            new THREE.Vector3(-w, h, d),
            new THREE.Vector3(0, 0, 0),
            new THREE.Vector3(w, -h, d),
            new THREE.Vector3(w, h, d),
            new THREE.Vector3(0, 0, 0),
            new THREE.Vector3(-w, -h, d),
            new THREE.Vector3(w, -h, d),
            new THREE.Vector3(0, 0, 0),
            new THREE.Vector3(-w, h, d),
            new THREE.Vector3(-w, -h, d),
        ]);

        var triSize = 0.2;
        var triGeometry = new THREE.Geometry();
        triGeometry.vertices.push(
            new THREE.Vector3(-triSize, h, d),
            new THREE.Vector3(0, h + triSize, d),
            new THREE.Vector3(triSize, h, d),
        );
        triGeometry.faces.push(new THREE.Face3(0, 1, 2));
        mainBody.add(new THREE.Mesh(triGeometry, ThreeGraph.Mat_Color()))

        mainBody.camInfo = {
            w: w,
            h: h,
            d: d,
        };

        return mainBody;
    },
    Cube: function Cube(size, material = ThreeGraph.Mat_Color()) {
        var s = size;
        var cube = ThreeGraph.Line([
            new THREE.Vector3(s, s, s),
            new THREE.Vector3(-s, s, s),
            new THREE.Vector3(-s, s, -s),
            new THREE.Vector3(s, s, -s),
            new THREE.Vector3(s, s, s),
            new THREE.Vector3(s, -s, s),
            new THREE.Vector3(-s, -s, s),
            new THREE.Vector3(-s, -s, -s),
            new THREE.Vector3(s, -s, -s),
            new THREE.Vector3(s, -s, s),
        ], material);

        cube.add(ThreeGraph.Line([
            new THREE.Vector3(-s, s, s),
            new THREE.Vector3(-s, -s, s),
        ], material));
        cube.add(ThreeGraph.Line([
            new THREE.Vector3(-s, s, -s),
            new THREE.Vector3(-s, -s, -s),
        ], material));
        cube.add(ThreeGraph.Line([
            new THREE.Vector3(s, s, -s),
            new THREE.Vector3(s, -s, -s),
        ], material));

        cube.tgPoints = [
            new THREE.Vector3(s, s, s),
            new THREE.Vector3(-s, s, s),
            new THREE.Vector3(-s, s, -s),
            new THREE.Vector3(s, s, -s),
            new THREE.Vector3(s, -s, s),
            new THREE.Vector3(-s, -s, s),
            new THREE.Vector3(-s, -s, -s),
            new THREE.Vector3(s, -s, -s),
        ];

        return cube;
    },

    RenderPoints: function RenderPoints(obj, material) {
        var pts = obj.tgPoints || [];

        pts.forEach(function(pos) {
            var geometry = new THREE.SphereGeometry(0.04, 12, 6);
            var point = new THREE.Mesh(geometry, material);
            point.isTgPoint = true;
            point.position.x = pos.x;
            point.position.y = pos.y;
            point.position.z = pos.z;
            point.originalPosition = new THREE.Vector3(pos.x, pos.y, pos.z);
            obj.add(point);
        });
    },

    Mat_Color: function(c = 0x000000) {
        var mat = new THREE.LineBasicMaterial({ color: c });
        mat.side = THREE.DoubleSide;
        mat.transparent = true;
        return mat;
    },
    Mat_ColorDashed: function(c = 0x000000, dashSize = 0.2, gapSize = 0.1) {
        var mat = new THREE.LineDashedMaterial({
           color: c,
           dashSize: dashSize,
           gapSize: gapSize,
        });
        mat.side = THREE.DoubleSide;
        mat.transparent = true;
        return mat;
    },

    Inverse: function Inverse(matrix) {
        var inverse = new THREE.Matrix4();
        inverse.getInverse(matrix);
        return inverse;
    },
    InverseMatrixWorld: function InverseMatrixWorld(obj) {
        obj.updateMatrix();
        obj.updateMatrixWorld(true);
        return ThreeGraph.Inverse(obj.matrixWorld);
    },
    ApplyLocalTransform: function ApplyLocalTransform(obj, matrix) {
        obj.updateMatrix();
        obj.updateMatrixWorld(true);
        obj.applyMatrix(matrix);
        return obj;
    },
    TransformVertices: function TransformVertices(obj, matrix) {
        ThreeGraph._forAllGeometries(obj, function(geom) {
            for (var i = 0; i < geom.vertices.length; i++) {
                geom.vertices[i].applyMatrix4(matrix);
            }
            geom.verticesNeedUpdate = true;
        });

        if (obj.tgPoints) {
            for (var i = 0; i < obj.tgPoints.length; i++) {
                obj.tgPoints[i].applyMatrix4(matrix);
            }
        }
    },

    TranslationMatrix: function TranslationMatrix(x, y, z) {
        if (x.x !== undefined && x.y !== undefined && x.z !== undefined) {
            // x is actually a vector
            y = x.y;
            z = x.z;
            x = x.x;
        }

        return new THREE.Matrix4().set(
            1,  0,  0,  x,
            0,  1,  0,  y,
            0,  0,  1,  z,
            0,  0,  0,  1,
        );
    },
    RotationXMatrix: function RotationXMatrix(rad) {
        return new THREE.Matrix4().set(
            1, 0, 0, 0,
            0, Math.cos(rad), -Math.sin(rad), 0,
            0, Math.sin(rad), Math.cos(rad), 0,
            0, 0, 0, 1,
        );
    },
    RotationYMatrix: function RotationYMatrix(rad) {
        return new THREE.Matrix4().set(
            Math.cos(rad), 0, Math.sin(rad), 0,
            0, 1, 0, 0,
            -Math.sin(rad), 0, Math.cos(rad), 0,
            0, 0, 0, 1,
        );
    },
    ProjectionMatrix: function ProjectionMatrix(fov) {
        var s = 1 / Math.tan(fov/2);

        return new THREE.Matrix4().set(
            s,  0,  0,  0,
            0,  s,  0,  0,
            0,  0,  0,  0,
            0,  0, -1,  0,
        );
    },

    _forAllGeometries: function _forAllGeometries(obj, func) {
        if (obj.geometry) {
            // obj has a geometry
            func(obj.geometry);
        }

        if (obj.children) {
            for (var i = 0; i < obj.children.length; i++) {
                _forAllGeometries(obj.children[i], func);
            }
        }
    }
};
