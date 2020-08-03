/*

Copyright (C) 2020  Daniele Rondina <geaaru@sabayonlinux.org>
Credits goes also to Gogs authors, some code portions and re-implemented design
are also coming from the Gogs project, which is using the go-macaron framework
and was really source of ispiration. Kudos to them!

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
package specs

type Client struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	File        string `json:"-" yaml:"-"`

	ActivitiesDirs []string `json:"activities_dirs,omitempty" yaml:"activities_dirs,omitempty"`

	Activities []Activity `json:"activities,omitempty" yaml:"activities,omitempty"`
}

type Scenario struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	File        string `json:"-" yaml:"-"`

	ResourceCosts []ResourceCost `json:"resources_cost,omitempty" yaml:"resources_cost,omitempty"`
	Rates         []ResourceRate `json:"rates,omitempty" yaml:"rates,omitempty"`

	NowTime string `json:"now" yaml:"name"`
}

type Activity struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Note        string `json:"note,omitempty" yaml:"note,omitempty"`
	Priority    int    `json:"priority" yaml:"priority"`
	File        string `json:"-" yaml:"-"`
	Disabled    bool   `json:"disabled,omitempty" yaml:"disabled,omitempty"`

	Tasks []Task `json:"tasks,omitempty" yaml:"tasks,omitempty"`
}

// General task structure for files specs
type Task struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Note        string `json:"note,omitempty" yaml:"note,omitempty"`
	Priority    int    `json:"priority" yaml:"priority"`
	Effort      string `json:"effort" yaml:"effort"`

	AllocatedResource []string `json:"resources,omitempty" yaml:"resources,omitempty"`

	Milestone string `json:"milestone,omitempty" yaml:"milestone,omitempty"`

	Flags  []string          `json:"flags,omitempty" yaml:"flags,omitempty"`
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	Tasks   []Task   `json:"subtasks,omitempty" yaml:"subtasks,omitempty"`
	Depends []string `json:"depends,omitempty" yaml:"depends,omitempty"`
}

type Resource struct {
	User  string `json:"user" yaml:"user"`
	Name  string `json:"name" yaml:"name"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
	Phone string `json:"phone,omitempty" yaml:"phone,omitempty"`
	File  string `json:"-" yaml:"-"`

	Holidays []ResourceHolidays `json:"holidays,omitempty" yaml:"holidays,omitempty"`
	Sick     []ResourceSick     `json:"sick,omitempty" yaml:"sick,omitempty"`
}

type Period struct {
	StartPeriod string `json:"start_period" yaml:"start_period"`
	EndPeriod   string `json:"end_period,omitempty" yaml:"end_period"`
}

type ResourceRate struct {
	*Period `json:"period" yaml:"period"`
	User    string  `json:"user" yaml:"user"`
	Rate    float64 `json:"rate" yaml:"rate"`
}

type ResourceCost struct {
	*Period `json:"period" yaml:"period"`
	User    string  `json:"user" yaml:"user"`
	Cost    float64 `json:"cost" yaml:"cost"`
}

type ResourceHolidays struct {
	*Period
}

type ResourceSick struct {
	*Period
}

type ResourceBooking struct {
	User string `json:"user" yaml:"user"`
}
