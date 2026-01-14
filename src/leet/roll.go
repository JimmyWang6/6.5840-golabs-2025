package leet

type Direction int

const (
	Left Direction = iota
	Right
	Up
	Down
)

type Point struct {
	x         int
	y         int
	direction Direction
}

// 给你一个 m 行 n 列的矩阵 matrix ，请按照 顺时针螺旋顺序 ，返回矩阵中的所有元素。
func spiralOrder(matrix [][]int) []int {
	m := len(matrix)
	if m == 0 {
		return []int{}
	}
	n := len(matrix[0])
	if n == 0 {
		return []int{}
	}
	visited := make([][]bool, m)
	for i := 0; i < m; i++ {
		visited[i] = make([]bool, n)
	}
	res := make([]int, 0)
	curPoint := Point{0, 0, Right}
	for {
		x, y := curPoint.x, curPoint.y
		res = append(res, matrix[x][y])
		visited[x][y] = true
		if len(res) >= m*n {
			break
		}
		curPoint = nextPoint(curPoint, visited)
	}
	return res
}

func nextPoint(currentPoint Point, visited [][]bool) Point {
	m := len(visited)
	n := len(visited[0])
	x, y := currentPoint.x, currentPoint.y
	direction := currentPoint.direction

	if direction == Up {
		if x-1 >= 0 && !visited[x-1][y] {
			x = x - 1
		} else {
			direction = nextDirection(direction)
			return nextPoint(Point{x, y, direction}, visited)
		}
	} else if direction == Down {
		if x+1 <= m-1 && !visited[x+1][y] {
			x = x + 1
		} else {
			direction = nextDirection(direction)
			return nextPoint(Point{x, y, direction}, visited)
		}
	} else if direction == Left {
		if y-1 >= 0 && !visited[x][y-1] {
			y = y - 1
		} else {
			direction = nextDirection(direction)
			return nextPoint(Point{x, y, direction}, visited)
		}
	} else if direction == Right {
		if y+1 <= n-1 && !visited[x][y+1] {
			y = y + 1
		} else {
			direction = nextDirection(direction)
			return nextPoint(Point{x, y, direction}, visited)
		}
	}
	return Point{x, y, direction}
}

func nextDirection(cur Direction) Direction {
	if cur == Left {
		return Up
	}
	if cur == Right {
		return Down
	}
	if cur == Down {
		return Left
	}
	return Right
}
