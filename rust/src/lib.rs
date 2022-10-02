extern crate js_sys;
extern crate wasm_bindgen;
extern crate rand;
use wasm_bindgen::prelude::*;

fn gen_range(low: usize, high: usize) -> usize {
    (js_sys::Math::random() * (high - low) as f64 + low as f64) as usize
}

#[derive(Copy, Clone, Debug)]
struct Coords {
    x: usize,
    y: usize,
}

struct Cell {
    connects_to: Vec<Coords>,
}

impl Cell {
    fn new() -> Cell {
        Cell {
            connects_to: Vec::with_capacity(4),
        }
    }
}

pub struct Maze {
    width: usize,
    height: usize,
    cells: Vec<Vec<Cell>>,
}

pub fn gen_maze(width: usize, height: usize) -> Maze {
    let mut cells: Vec<Vec<Cell>> = Vec::with_capacity(width);
    for _ in 0..width {
        let mut column: Vec<Cell> = Vec::with_capacity(height);
        for _ in 0..height {
            column.push(Cell::new());
        }

        cells.push(column);
    }

    let mut visited = vec![vec![false; height]; width];

    let start = Coords { x: gen_range(0, width), y: gen_range(0, height) };
    let mut stack = vec![start];

    while stack.len() > 0 {
        let c = *stack.last().unwrap();

        visited[c.x][c.y] = true;

        let mut unvisited_neighbors: Vec<Coords> = Vec::with_capacity(4);
        let directions: Vec<Vec<isize>> = vec![vec![-1, 0], vec![1, 0], vec![0, -1], vec![0, 1]];
        for d in directions {
            let new_x = (c.x as isize + d[0]) as usize;
            let new_y = (c.y as isize + d[1]) as usize;

            if new_x < 0 || new_x >= width || new_y < 0 || new_y >= height || visited[new_x][new_y] {
                continue;
            } else {
                unvisited_neighbors.push(Coords{x: new_x, y: new_y});
            }
        }

        if unvisited_neighbors.len() > 0 {
            let n = unvisited_neighbors[gen_range(0, unvisited_neighbors.len())];
            
            let current_cell = &mut cells[c.x][c.y];
            current_cell.connects_to.push(n);

            let neighbor_cell = &mut cells[n.x][n.y];
            neighbor_cell.connects_to.push(c);

            stack.push(n);
        } else {
            stack.pop();
        }
    }

    Maze {
        width: width,
        height: height,
        cells: cells,
    }
}

#[wasm_bindgen]
pub fn gen_maze_rust_silent(width: usize, height: usize) {
    gen_maze(width, height);
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn it_works() {
        gen_maze(10, 5);
    }
}

