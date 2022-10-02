/*

Data structure of a maze:

{
  width: num,
  height: num,
  cells: [ // columns, left to right
    [ // rows, top to bottom
      {
        connectsTo: [
          [x, y],
          // ...
        ],
      },
      // ...
    ]
  ]
}
 */

function randomInt(min, max) {
  min = Math.ceil(min);
  max = Math.floor(max);
  return Math.floor(Math.random() * (max - min)) + min;
}

function genArray(width, height, valueFunc) {
  const result = new Array(width);
  for (let x = 0; x < width; x++) {
    result[x] = new Array(height);

    for (let y = 0; y < height; y++) {
      result[x][y] = valueFunc();
    }
  }

  return result;
}

function genMazeJS(width, height, callback) {
  const cells = genArray(width, height, () => ({ connectsTo: [] }));

  const start = [randomInt(0, width), randomInt(0, height)];
  
  const stack = [];
  const visited = genArray(width, height, () => false);
  stack.push(start);

  while (stack.length > 0) {
    const currentCoords = stack[stack.length - 1];
    const [x, y] = currentCoords;

    const currentCell = cells[x][y];

    visited[x][y] = true;

    const unvisitedNeighbors = [];
    for (const d of [[-1, 0], [1, 0], [0, -1], [0, 1]]) {
      const newX = x + d[0];
      const newY = y + d[1];
      if (
        newX < 0
        || newX >= width
        || newY < 0
        || newY >= height
        || visited[newX][newY]
      ) {
        continue;
      } else {
        unvisitedNeighbors.push([newX, newY]);
      }
    }

    if (unvisitedNeighbors.length > 0) {
      const chosenNeighborCoords = unvisitedNeighbors[randomInt(0, unvisitedNeighbors.length)];
      const [nx, ny] = chosenNeighborCoords;
      currentCell.connectsTo.push([nx, ny]);

      const neighborCell = cells[nx][ny];
      neighborCell.connectsTo.push([x, y]);

      stack.push([nx, ny]);
    } else {
      stack.pop();
    }
  }

  callback({
    width,
    height,
    cells,
  });
}

function printMazeOne(maze) {
  let result = '';
  for (let y = 0; y < maze.height; y++) {
    for (let x = 0; x < maze.width; x++) {
      const cell = maze.cells[x][y];
      let left = false, right = false, up = false, down = false;

      for (const [nx, ny] of cell.connectsTo) {
        if (nx < x) left  = true;
        if (nx > x) right = true;
        if (ny < y) up    = true;
        if (ny > y) down  = true;
      }

      let vert = ' ';
      if (!up && !down) {
        vert = '=';
      } else if (!up) {
        vert = '‾';
      } else if (!down) {
        vert = '_';
      }

      result += (left ? ' ' : '|') + vert + (right ? ' ' : '|');
    }
    result += '\n';
  }

  console.log(result);
}

function mazeToText(maze) {
  let result = '';
  for (let y = 0; y < maze.height; y++) {
    const cellDirs = [];

    for (let x = 0; x < maze.width; x++) {
      const cell = maze.cells[x][y];
      let left = false, right = false, up = false, down = false;

      for (const [nx, ny] of cell.connectsTo) {
        if (nx < x) left  = true;
        if (nx > x) right = true;
        if (ny < y) up    = true;
        if (ny > y) down  = true;
      }

      cellDirs.push([left, right, up, down]);
    }

    for (const [left, right, up, down] of cellDirs) {
      result += '┼' + (up ? ' ' : '─');
    }
    result += '┼\n';

    for (const [left, right, up, down] of cellDirs) {
      result += (left ? ' ' : '│') + ' ';
    }
    result += '│\n';
  }

  for (let x = 0; x < maze.width; x++) {
    result += '┼─';
  }
  result += '┼';

  return result;
}
