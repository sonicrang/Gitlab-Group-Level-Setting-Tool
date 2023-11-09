package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

var LstProjects = []Project{}
var URL, Token string

// 预设Label
var LstLabels = []Label{
	{
		Name:  "QC::并发多线程",
		Color: "#009966",
	},
	{
		Name:  "QC::服务接口定义",
		Color: "#8fbc8f",
	},
	{
		Name:  "QC::健壮性",
		Color: "#013220",
	},
	{
		Name:  "QC::日志",
		Color: "#6699cc",
	},
	{
		Name:  "QC::性能",
		Color: "#0000ff",
	},
	{
		Name:  "QC::异常",
		Color: "#e6e6fa",
	},
	{
		Name:  "QC::规范类",
		Color: "#9400d3",
	},
	{
		Name:  "QC::安全编码",
		Color: "#ff0000",
	},
	{
		Name:  "QC::其他",
		Color: "#eee600",
	},
}

type Label struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Project struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Branch struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Group struct {
	ID int `json:"id"`
}

type Rule struct {
	Name      string `json:"name"`
	Type      string `json:"rule_type"`
	Require   int    `json:"approvals_required"`
	GroupIDs  []int  `json:"group_ids"`
	BranchIDs []int  `json:"protected_branch_ids"`
}

type PushRule struct {
	commit_committer_check bool `json:"commit_committer_check"`
	// TODO filed not exist in response
	commit_committer_name_check bool `json:"commit_committer_name_check"`
	reject_unsigned_commits     bool `json:"reject_unsigned_commits"`
	// TODO filed not exist in response
	reject_non_dco_commits        bool   `json:"reject_non_dco_commits"`
	deny_delete_tag               bool   `json:"deny_delete_tag"`
	member_check                  bool   `json:"member_check"`
	prevent_secrets               bool   `json:"prevent_secrets"`
	commit_message_regex          string `json:"commit_message_regex"`
	commit_message_negative_regex string `json:"commit_message_negative_regex"`
	branch_name_regex             string `json:"branch_name_regex"`
	author_email_regex            string `json:"author_email_regex"`
	file_name_regex               string `json:"file_name_regex"`
	max_file_size                 int    `json:"max_file_size"`
}

func main() {
	fmt.Print("\033[H\033[2J")
	fmt.Println("---------- 初始化设置 ----------")
	fmt.Println("a. 请输入GitLab URL，如：https://jihulab.com")
	fmt.Scanln(&URL)
	fmt.Println("b. 请输入GitLab Access Token")
	fmt.Scanln(&Token)
	SetGroupID()
}

func SetGroupID() {
	LstProjects = []Project{}
	var group_id int
	fmt.Print("\033[H\033[2J")
	fmt.Println("---------- 群组设置 ----------")
	fmt.Println("a. 请输入群组ID")
	fmt.Scanln(&group_id)
	GetProjectsByGroupID(group_id)
	fmt.Println("该群组下的项目有【" + strconv.Itoa(len(LstProjects)) + "】个：")
	for _, v := range LstProjects {
		fmt.Println(strconv.Itoa(v.ID) + " : " + v.Name)
	}
	fmt.Println("按任意键继续")
	fmt.Scanln()
	ShowMenu(group_id)
}

func ShowMenu(group_id int) {
	var protect_branch_name, approval_rule_name string
	var push_access_level, merge_access_level, approval_required int
	var approval_group_ids []int

	for true {
		fmt.Print("\033[H\033[2J")
		fmt.Println("【群组】 " + strconv.Itoa(group_id))
		fmt.Println("---------- 菜单 ----------")
		fmt.Println("1. 设置保护分支")
		fmt.Println("2. 设置审批规则")
		fmt.Println("3. 设置合并检查：需解决所有讨论")
		fmt.Println("4. 设置Label")
		fmt.Println("5. 下发群组推送规则到项目")
		fmt.Println("98. 返回上一级")
		fmt.Println("99. 退出")
		fmt.Println("\n请输入功能编号：")
		var choose int
		fmt.Scanln(&choose)
		switch choose {
		case 1:
			fmt.Println("a. 请输入保护分支名称或规则，如main、*-stable")
			fmt.Scanln(&protect_branch_name)
			fmt.Println("b. 请输入允许推送的人员类型：0 => No access, 30 => Developer access, 40 => Maintainer access, 60 => Admin access")
			fmt.Scanln(&push_access_level)
			fmt.Println("c. 请输入允许合并的人员类型：0 => No access, 30 => Developer access, 40 => Maintainer access, 60 => Admin access")
			fmt.Scanln(&merge_access_level)
			SetProtectBranch(protect_branch_name, push_access_level, merge_access_level)
		case 2:
			fmt.Println("a. 请输入审批规则名称，如开发组、审核组")
			fmt.Scanln(&approval_rule_name)
			fmt.Println("b. 请输入审批组ID")
			var approval_group_id int
			fmt.Scanln(&approval_group_id)
			approval_group_ids = append(approval_group_ids, approval_group_id)
			fmt.Println("c. 请输入最小核准人数")
			fmt.Scanln(&approval_required)
			fmt.Println("d. 请输入规则生效的保护分支名,如main、*-stable，留空则为所有分支")
			protect_branch_name = ""
			fmt.Scanln(&protect_branch_name)
			SetApprovalRules(approval_rule_name, approval_group_ids, approval_required, protect_branch_name)
		case 3:
		INPUT:
			fmt.Println("a. 是否开启合并检查：需解决所有讨论（true/false）")
			var isAllTreadResolved string
			fmt.Scanln(&isAllTreadResolved)
			if isAllTreadResolved == "true" || isAllTreadResolved == "false" {
				SetMergeCheck_AllTreadResolved(isAllTreadResolved)
			} else {
				goto INPUT
			}
		case 4:
			SetLabels(group_id, LstLabels)
		case 5:
			SyncGroupPushRuleToProjects(group_id)
		case 98:
			SetGroupID()
		case 99:
			os.Exit(0)
		}
		fmt.Println("按任意键返回")
		fmt.Scanln()
	}
}

func GetProjectsByGroupID(group_id int) {
	client := &http.Client{}

	// Get SubGroups
	lstGroups := []Group{}
	page := 1
GroupPagination:
	req, _ := http.NewRequest("GET", URL+"/api/v4/groups/"+strconv.Itoa(group_id)+"/subgroups?per_page=100&page="+strconv.Itoa(page), nil)
	req.Header.Add("PRIVATE-TOKEN", Token)
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	groups := []Group{}
	err := json.Unmarshal(body, &groups)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(groups) > 0 {
		page++
		lstGroups = append([]Group{}, groups...)
		goto GroupPagination
	}
	for _, v := range lstGroups {
		// Recursion
		GetProjectsByGroupID(v.ID)
	}

	// Get Projects
	page = 1
ProjectPagination:
	req, _ = http.NewRequest("GET", URL+"/api/v4/groups/"+strconv.Itoa(group_id)+"/projects?per_page=100&page="+strconv.Itoa(page), nil)
	req.Header.Add("PRIVATE-TOKEN", Token)
	resp, _ = client.Do(req)
	body, _ = ioutil.ReadAll(resp.Body)
	projects := []Project{}
	err = json.Unmarshal(body, &projects)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(projects) > 0 {
		page++
		LstProjects = append(LstProjects, projects...)
		goto ProjectPagination
	}
}

func SetProtectBranch(protect_branch_name string, push_access_level int, merge_access_level int) {
	client := &http.Client{}
	for _, v := range LstProjects {
		req, _ := http.NewRequest("POST", URL+"/api/v4/projects/"+strconv.Itoa(v.ID)+
			"/protected_branches?name="+protect_branch_name+"&push_access_level="+strconv.Itoa(push_access_level)+
			"&merge_access_level="+strconv.Itoa(merge_access_level), nil)
		req.Header.Add("PRIVATE-TOKEN", Token)
		resp, _ := client.Do(req)

		if resp.StatusCode != 201 {
			fmt.Println("设置群组保护分支失败！")
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("StatusCode: " + strconv.Itoa(resp.StatusCode) + "\nBody: " + string(body))
			return
		}
	}
	fmt.Println("设置群组保护分支成功！")
}

func SetApprovalRules(approval_rule_name string, approval_group_ids []int, approval_required int, protect_branch_name string) {
	client := &http.Client{}
	for _, v := range LstProjects {
		var protected_branch_ids []int
		if protect_branch_name != "" {
			// Get protected branches
			req, _ := http.NewRequest("GET", URL+"/api/v4/projects/"+strconv.Itoa(v.ID)+"/protected_branches", nil)
			req.Header.Add("PRIVATE-TOKEN", Token)
			resp, _ := client.Do(req)
			body, _ := ioutil.ReadAll(resp.Body)

			branches := []Branch{}
			err := json.Unmarshal(body, &branches)
			if err != nil {
				fmt.Println(err)
				return
			}
			for _, b := range branches {
				if b.Name == protect_branch_name {
					protected_branch_ids = append(protected_branch_ids, b.ID)
					break
				}
			}
		}
		// SetApprovalRules
		rule := Rule{
			Name:      approval_rule_name,
			Type:      "regular",
			Require:   approval_required,
			GroupIDs:  approval_group_ids,
			BranchIDs: protected_branch_ids,
		}
		json, err := json.Marshal(rule)
		if err != nil {
			fmt.Println(err)
			return
		}
		req, _ := http.NewRequest("POST", URL+"/api/v4/projects/"+strconv.Itoa(v.ID)+
			"/approval_rules", bytes.NewBuffer(json))
		req.Header.Add("PRIVATE-TOKEN", Token)
		req.Header.Add("Content-Type", "application/json")
		resp, _ := client.Do(req)

		if resp.StatusCode != 201 {
			fmt.Println("设置群组审批规则失败！")
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("StatusCode: " + strconv.Itoa(resp.StatusCode) + "\nBody: " + string(body))
			return
		}
	}
	fmt.Println("设置群组审批规则成功！")
}

func SetMergeCheck_AllTreadResolved(isAllTreadResolved string) {
	client := &http.Client{}
	for _, v := range LstProjects {
		req, _ := http.NewRequest("PUT", URL+"/api/v4/projects/"+strconv.Itoa(v.ID)+"?only_allow_merge_if_all_discussions_are_resolved="+isAllTreadResolved, nil)
		req.Header.Add("PRIVATE-TOKEN", Token)
		resp, _ := client.Do(req)

		if resp.StatusCode != 200 {
			fmt.Println("设置群组合并检查失败！")
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("StatusCode: " + strconv.Itoa(resp.StatusCode) + "\nBody: " + string(body))
			return
		}
	}
	fmt.Println("设置群组合并检查成功！")

}

func SetLabels(group_id int, LstLabels []Label) {
	client := &http.Client{}

	for _, v := range LstLabels {
		json, err := json.Marshal(v)
		if err != nil {
			fmt.Println(err)
			return
		}
		req, _ := http.NewRequest("POST", URL+"/api/v4/groups/"+strconv.Itoa(group_id)+"/labels", bytes.NewBuffer(json))
		req.Header.Add("PRIVATE-TOKEN", Token)
		req.Header.Add("Content-Type", "application/json")
		resp, _ := client.Do(req)

		if resp.StatusCode != 201 {
			fmt.Println("设置群组Label失败！")
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("StatusCode: " + strconv.Itoa(resp.StatusCode) + "\nBody: " + string(body))
			return
		}
	}
	fmt.Println("设置群组Label成功！")

}

func SyncGroupPushRuleToProjects(group_id int) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", URL+"/api/v4/groups/"+strconv.Itoa(group_id)+"/push_rule", nil)
	req.Header.Add("PRIVATE-TOKEN", Token)
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	pushRule := PushRule{}
	err := json.Unmarshal(body, &pushRule)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(pushRule.reject_unsigned_commits)
	pushRule.reject_unsigned_commits = true
	fmt.Println(pushRule.reject_unsigned_commits)

	fmt.Println(pushRule.commit_message_regex)
	for _, v := range LstProjects {
		fmt.Println("项目" + v.Name + "设置中...")
		req1, _ := http.NewRequest("GET", URL+"/api/v4/projects/"+strconv.Itoa(v.ID)+"/push_rule", bytes.NewBuffer(body))
		req1.Header.Add("PRIVATE-TOKEN", Token)
		req1.Header.Set("Content-Type", "application/json")
		resp1, _ := client.Do(req1)
		body1, _ := ioutil.ReadAll(resp1.Body)
		var httpMethod = "PUT"
		// 如果项目没有设置推送规则，则使用POST进行创建
		if string(body1) == "null" {
			httpMethod = "POST"
		}

		req, _ := http.NewRequest(httpMethod, URL+"/api/v4/projects/"+strconv.Itoa(v.ID)+"/push_rule", bytes.NewBuffer(body))
		req.Header.Add("PRIVATE-TOKEN", Token)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := client.Do(req)

		// 现场发现返回了 200、201
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			fmt.Println("设置失败！")
			errBody, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("StatusCode: " + strconv.Itoa(resp.StatusCode) + "\nBody: " + string(errBody))
			return
		}
	}
	fmt.Println("设置成功！")
}
