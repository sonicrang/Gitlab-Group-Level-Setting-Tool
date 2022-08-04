package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

var lstProjects = []Project{}

func main() {
	var url, token string
	fmt.Println("---------- 初始化设置 ----------")
	fmt.Println("a. 请输入GitLab URL，如：https://jihulab.com")
	fmt.Scanln(&url)

	fmt.Println("b. 请输入GitLab Access Token")
	fmt.Scanln(&token)
	ShowMenu(url, token)
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

func ShowMenu(url, token string) {
	var protect_branch_name, approval_rule_name string
	var group_id, push_access_level, merge_access_level, approval_required int
	var approval_group_ids []int

	for true {
		lstProjects = []Project{}
		fmt.Print("\033[H\033[2J")
		fmt.Println("---------- 菜单 ----------")
		fmt.Println("1. 设置群组保护分支")
		fmt.Println("2. 设置群组审批规则")
		fmt.Println("3. 设置群组合并检查：需解决所有讨论")
		fmt.Println("4. 退出")
		fmt.Println("\n请输入功能编号：")
		var choose int
		fmt.Scanln(&choose)
		switch choose {
		case 1:
			fmt.Println("a. 请输入需要进行设置的群组ID")
			fmt.Scanln(&group_id)
			GetProjectsByGroupID(url, token, group_id)
			fmt.Println("该群组下的项目有：")
			for _, v := range lstProjects {
				fmt.Println(strconv.Itoa(v.ID) + " : " + v.Name)
			}
			fmt.Println("b. 请输入保护分支名称或规则，如main、*-stable")
			fmt.Scanln(&protect_branch_name)
			fmt.Println("c. 请输入允许推送的人员类型：0 => No access, 30 => Developer access, 40 => Maintainer access, 60 => Admin access")
			fmt.Scanln(&push_access_level)
			fmt.Println("d. 请输入允许合并的人员类型：0 => No access, 30 => Developer access, 40 => Maintainer access, 60 => Admin access")
			fmt.Scanln(&merge_access_level)
			SetProtectBranch(url, token, protect_branch_name, push_access_level, merge_access_level)
		case 2:
			fmt.Println("a. 请输入需要进行设置的群组ID")
			fmt.Scanln(&group_id)
			GetProjectsByGroupID(url, token, group_id)
			fmt.Println("该群组下的项目有：")
			for _, v := range lstProjects {
				fmt.Println(strconv.Itoa(v.ID) + " : " + v.Name)
			}
			fmt.Println("b. 请输入审批规则名称，如开发组、审核组")
			fmt.Scanln(&approval_rule_name)
			fmt.Println("c. 请输入审批组ID")
			var approval_group_id int
			fmt.Scanln(&approval_group_id)
			approval_group_ids = append(approval_group_ids, approval_group_id)
			fmt.Println("d. 请输入最小核准人数")
			fmt.Scanln(&approval_required)
			fmt.Println("e. 请输入规则生效的保护分支名,如main、*-stable，留空则为所有分支")
			protect_branch_name = ""
			fmt.Scanln(&protect_branch_name)
			SetApprovalRules(url, token, approval_rule_name, approval_group_ids, approval_required, protect_branch_name)
		case 3:
			fmt.Println("a. 请输入需要进行设置的群组ID")
			fmt.Scanln(&group_id)
			GetProjectsByGroupID(url, token, group_id)
			fmt.Println("该群组下的项目有：")
			for _, v := range lstProjects {
				fmt.Println(strconv.Itoa(v.ID) + " : " + v.Name)
			}
		INPUT:
			fmt.Println("b. 是否开启合并检查：需解决所有讨论（true/false）")
			var isAllTreadResolved string
			fmt.Scanln(&isAllTreadResolved)
			if isAllTreadResolved == "true" || isAllTreadResolved == "false" {
				SetMergeCheck_AllTreadResolved(url, token, isAllTreadResolved)
			} else {
				goto INPUT
			}
		case 4:
			return
		}
		fmt.Println("按任意键返回")
		fmt.Scanln()
	}
}

func GetProjectsByGroupID(url string, token string, group_id int) {
	client := &http.Client{}

	// Get SubGroups
	lstGroups := []Group{}
	page := 1
GroupPagination:
	req, _ := http.NewRequest("GET", url+"/api/v4/groups/"+strconv.Itoa(group_id)+"/subgroups?per_page=100&page="+strconv.Itoa(page), nil)
	req.Header.Add("PRIVATE-TOKEN", token)
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
		GetProjectsByGroupID(url, token, v.ID)
	}

	// Get Projects
	page = 1
ProjectPagination:
	req, _ = http.NewRequest("GET", url+"/api/v4/groups/"+strconv.Itoa(group_id)+"/projects?per_page=100&page="+strconv.Itoa(page), nil)
	req.Header.Add("PRIVATE-TOKEN", token)
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
		lstProjects = append(lstProjects, projects...)
		goto ProjectPagination
	}
}

func SetProtectBranch(url string, token string, protect_branch_name string, push_access_level int, merge_access_level int) {
	client := &http.Client{}
	for _, v := range lstProjects {
		req, _ := http.NewRequest("POST", url+"/api/v4/projects/"+strconv.Itoa(v.ID)+
			"/protected_branches?name="+protect_branch_name+"&push_access_level="+strconv.Itoa(push_access_level)+
			"&merge_access_level="+strconv.Itoa(merge_access_level), nil)
		req.Header.Add("PRIVATE-TOKEN", token)
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

func SetApprovalRules(url string, token string, approval_rule_name string, approval_group_ids []int, approval_required int, protect_branch_name string) {
	client := &http.Client{}
	for _, v := range lstProjects {
		var protected_branch_ids []int
		if protect_branch_name != "" {
			// Get protected branches
			req, _ := http.NewRequest("GET", url+"/api/v4/projects/"+strconv.Itoa(v.ID)+"/protected_branches", nil)
			req.Header.Add("PRIVATE-TOKEN", token)
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
		req, _ := http.NewRequest("POST", url+"/api/v4/projects/"+strconv.Itoa(v.ID)+
			"/approval_rules", bytes.NewBuffer(json))
		req.Header.Add("PRIVATE-TOKEN", token)
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

func SetMergeCheck_AllTreadResolved(url, token, isAllTreadResolved string) {
	client := &http.Client{}
	for _, v := range lstProjects {
		req, _ := http.NewRequest("PUT", url+"/api/v4/projects/"+strconv.Itoa(v.ID)+"?only_allow_merge_if_all_discussions_are_resolved="+isAllTreadResolved, nil)
		req.Header.Add("PRIVATE-TOKEN", token)
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
