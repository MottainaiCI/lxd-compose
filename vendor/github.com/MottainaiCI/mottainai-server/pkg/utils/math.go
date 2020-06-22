/*

Copyright (C) 2018  Ettore Di Giacinto <mudler@gentoo.org>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.

*/

package utils

func FeatureScaling(i, imax, min, max float64) float64 {
	return min + ((i-1)*(max-min))/(imax-1)
}

func LogisticMap(r, xn float64) float64 {
	return r * xn * (1 - xn)
}

func LogisticMapSteps(steps int, r, xn float64) float64 {

	if r < -2.0 {
		r = -2.0
	}

	if r > 4.0 {
		r = 4.0
	}

	if xn <= 0 {
		xn = 0.1
	}

	if xn > 1.0 { // Avoid -inf divergence
		xn = 1.0
	}

	for i := 0; i < steps; i++ {
		xn = LogisticMap(r, xn)
	}

	return xn
}
