package leet

func maxArea(height []int) int {
	maxHeight := 0
	lp := 0
	rp := len(height) - 1
	for lp < rp {
		if height[lp] < height[rp] {
			if ((rp - lp) * height[lp]) > maxHeight {
				maxHeight = height[lp] * (rp - lp)
			}
			lp++
		} else {
			if ((rp - lp) * height[rp]) > maxHeight {
				maxHeight = height[rp] * (rp - lp)
			}
			rp--
		}
	}
	return maxHeight
}
