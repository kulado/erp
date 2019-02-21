package controllers

import (
	"html/template"
	"strconv"

	"github.com/kulado/erp/models"
	"github.com/kulado/erp/plugins/permission"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/xuri/excelize"
)

type CategoryController struct {
	beego.Controller
}

// Category list data front display, Categories list data show reception
func (c *CategoryController) Get() {
	o := orm.NewOrm()
	category := []models.Category{}
	o.QueryTable("category").Filter("is_hidden", false).OrderBy("primary", "two_stage").All(&category)

	if len(category) == 0 {
		c.Data["msg"] = "The classification table data is empty, please contact the administrator to add classification information~, Classification data is empty, contact your administrator to add classifieds -"
		c.Data["url"] = "/"
		c.TplName = "jump/error.html"
		return
	}

	// i, j are the number of primary and secondary classifications respectively, i, j are a number of categories and sub-categories
	var i, j int64 = 0, 0
	for _, item := range category {
		if item.TwoStage == "-" {
			i++
		}
		if item.TwoStage != "-" && item.ThreeStage == "-" {
			j++
		}
	}

	//primary, two_stage is a map of the primary and secondary classifications with the database Id as the index and the classification name as the value., two_stage Id respectively in the database as an index value to the category names map classification and a classification of the two
	primary := make(map[int]string, i)
	two_stage := make(map[int]string, j)
	for _, item := range category {
		if item.TwoStage == "-" {
			primary[item.Id] = item.Primary
		}
		if item.TwoStage != "-" && item.ThreeStage == "-" {
			two_stage[item.Id] = item.TwoStage
		}
	}

	c.Data["category"] = category
	c.Data["primary"] = primary
	c.Data["two_stage"] = two_stage
	c.Layout = "common.tpl"
	c.TplName = "category/category_list.html"
}

// Submit classification table excel interface display, Submit classification excel interface display
func (c *CategoryController) Category_upload() {
	if !permission.GetOneItemPermission(c.GetSession("username").(string), "OperateCategory") {
		c.Abort("401")
	}
	c.Layout = "common.tpl"
	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML())
	c.TplName = "category/category_upload.html"
}

// Classification table excel file upload, and update database classification table, Classification excel file uploads, and update the database classification
func (c *CategoryController) Category_upload_post() {
	if !permission.GetOneItemPermission(c.GetSession("username").(string), "OperateCategory") {
		c.Abort("401")
	}
	f, h, err := c.GetFile("category_file")
	if err != nil {
		logs.Error("User ID：", c.GetSession("uid"), "Uploading category_file failed, reason:, Upload category_file failed because:", err)
		c.Data["msg"] = "Upload file error, please check and try again~, Upload file error. Please check retry ~"
		c.Data["url"] = "/category_upload"
		c.TplName = "jump/error.html"
	}
	defer f.Close()
	if h.Header.Get("Content-Type") == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		// File Upload, Upload file
		filename := strconv.Itoa(c.GetSession("uid").(int)) + "_" + h.Filename
		c.SaveToFile("category_file", "static/upload/"+filename)

		// Xlsx file parsing
		xlFile, err := excelize.OpenFile("./static/upload/" + filename)
		if err != nil {
			logs.Error("360EntSecGroup-Skylar/excelize: reading xlsx file failed->", err)
			c.Data["msg"] = "Failed to read .xlsx file, please try and try again~, Read .xlsx file failed, please try again check ~"
			c.Data["url"] = "/category_upload"
			c.TplName = "jump/error.html"
			return
		}

		rows := xlFile.GetRows("sheet2")
		rowsnum := len(rows)
		catagory := make([]models.Category, rowsnum)
		var i, j int = 0, 0
		for _, row := range rows {
			temp := make([]string, 5)
			j = 0
			for _, colCell := range row {
				temp[j] = colCell
				j++
			}

			// Convert Id to int type, The converted to type int Id
			id, _ := strconv.Atoi(temp[0])

			// Said is_hidden converted to bool type value, Speak is_hidden converted to bool type value
			var is_hidden bool = false
			if temp[4] == "1" {
				is_hidden = true
			}

			catagory[i] = models.Category{Id: id, Primary: temp[1], TwoStage: temp[2], ThreeStage: temp[3], Is_hidden: is_hidden}

			i++
		}
		catagory = catagory[1:]

		// Database operation, Database operations
		o := orm.NewOrm()
		o.Raw("truncate table category").Exec()
		nums, err := o.InsertMulti(100, catagory)
		if err != nil {
			logs.Error(err)
		} else {
			c.Data["msg"] = "Database classification table co-insertion, Were inserted into the database classification" + strconv.Itoa(int(nums)) + "Article data~, Of data ~"
			c.Data["url"] = "/category_list"
			c.TplName = "jump/success.html"
			return
		}

	} else {
		c.Data["msg"] = "Please upload the excel file with extension .xlsx~, Please upload .xlsx extension of the excel file ~"
		c.Data["url"] = "/category_upload"
		c.TplName = "jump/error.html"
	}
}

// add category
func (c *CategoryController) Category_add() {
	if !permission.GetOneItemPermission(c.GetSession("username").(string), "OperateCategory") {
		c.Abort("401")
	}
	category := []models.Category{}
	o := orm.NewOrm()

	// Query data once, Check out the one-time data
	o.QueryTable("category").Filter("three_stage", "-").All(&category)

	var primary_string string
	var two_stage_string string
	for _, item := range category {
		if item.TwoStage == "-" {
			primary_string += item.Primary + ", "
		} else {
			two_stage_string += item.TwoStage + ", "
		}
	}

	c.Data["primary_string"] = primary_string
	c.Data["two_stage_string"] = two_stage_string
	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML())
	c.Layout = "common.tpl"
	c.TplName = "category/category_add.html"
}

// Add category submission, Add categories to submit
func (c *CategoryController) Category_add_post() {
	if !permission.GetOneItemPermission(c.GetSession("username").(string), "OperateCategory") {
		c.Abort("401")
	}
	primary := c.GetString("primary")
	two_stage := c.GetString("two_stage")
	three_stage := c.GetString("three_stage")

	category := models.Category{}
	category.Primary = primary
	category.TwoStage = two_stage
	category.ThreeStage = three_stage

	o := orm.NewOrm()
	primary_query := models.Category{}
	o.QueryTable("category").Filter("primary", primary).
		One(&primary_query, "id", "is_hidden")
	// If the query is not found, it will return the zero value of the corresponding type., If the query returns less than zero value corresponding to the type of
	if primary_query.Id != 0 {
		category.Primary = strconv.Itoa(primary_query.Id)
	}

	two_stage_query := models.Category{}
	o.QueryTable("category").Filter("two_stage", two_stage).
		One(&two_stage_query, "id", "is_hidden")
	// If the query is not found, it will return the zero value of the corresponding type., If the query returns less than zero value corresponding to the type of
	if two_stage_query.Id != 0 {
		category.TwoStage = strconv.Itoa(two_stage_query.Id)
	}

	// Determine whether to hide, To determine whether hidden
	if primary_query.Is_hidden || two_stage_query.Is_hidden {
		category.Is_hidden = true
	} else {
		category.Is_hidden = false
	}

	_, err := o.Insert(&category)
	if err != nil {
		c.Data["url"] = "/category_add"
		c.Data["msg"] = "Add category failed, Add categories failure~"
		c.TplName = "jump/error.html"
		return
	} else {
		c.Data["url"] = "/category_list"
		c.Data["msg"] = "Add classification successfully, Add categories success~"
		c.TplName = "jump/success.html"
	}
}

// Category editing, Category editor
func (c *CategoryController) Category_edit() {
	if !permission.GetOneItemPermission(c.GetSession("username").(string), "OperateCategory") {
		c.Abort("401")
	}
	category := []models.Category{}
	o := orm.NewOrm()
	// Here only the search convenience is added to the commonly used three-level classification., Here only add convenience to retrieve the common three-tier classification
	o.QueryTable("category").Exclude("three_stage", "-").All(&category, "three_stage")

	var three_stage_string string
	for _, item := range category {
		three_stage_string += item.ThreeStage + ", "
	}

	c.Data["three_stage_string"] = three_stage_string
	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML())
	c.Layout = "common.tpl"
	c.TplName = "category/category_edit.html"
}

// Ajax requests 1 category information, ajax request a classification information
func (c *CategoryController) Category_search() {
	if !c.IsAjax() {
		return
	}
	if !permission.GetOneItemPermission(c.GetSession("username").(string), "OperateCategory") {
		c.Abort("401")
	}

	o := orm.NewOrm()
	category := models.Category{}
	primary := models.Category{}
	two_stage := models.Category{}
	three_stage := models.Category{}

	item := c.GetString("item")
	stage := c.GetString("stage")

	switch stage {
	case "primary":
		o.QueryTable("category").Filter("primary", item).One(&primary)
		o.QueryTable("category").Filter("id", primary.TwoStage).One(&two_stage, "two_stage")
		o.QueryTable("category").Filter("id", primary.ThreeStage).One(&three_stage, "three_stage")
		category.Id = primary.Id
		category.Is_hidden = primary.Is_hidden
	case "two_stage":
		o.QueryTable("category").Filter("two_stage", item).One(&two_stage)
		o.QueryTable("category").Filter("id", two_stage.Primary).One(&primary, "primary")
		o.QueryTable("category").Filter("id", two_stage.ThreeStage).One(&three_stage, "three_stage")
		category.Id = two_stage.Id
		category.Is_hidden = two_stage.Is_hidden
	case "three_stage":
		o.QueryTable("category").Filter("three_stage", item).One(&three_stage)
		o.QueryTable("category").Filter("id", three_stage.TwoStage).One(&two_stage, "two_stage")
		o.QueryTable("category").Filter("id", three_stage.Primary).One(&primary, "primary")
		category.Id = three_stage.Id
		category.Is_hidden = three_stage.Is_hidden
	}

	category.Primary = primary.Primary
	category.TwoStage = two_stage.TwoStage
	category.ThreeStage = three_stage.ThreeStage
	c.Data["json"] = category
	c.ServeJSON()
}

// Category editing post, Categories edit post
func (c *CategoryController) Category_edit_post() {
	if !permission.GetOneItemPermission(c.GetSession("username").(string), "OperateCategory") {
		c.Abort("401")
	}

	category := models.Category{}
	category.Id, _ = c.GetInt("category_id")
	category.Primary = c.GetString("primary")
	category.TwoStage = c.GetString("two_stage")
	category.ThreeStage = c.GetString("three_stage")

	if category.Primary == "-" {
		c.Data["url"] = "/category_edit"
		c.Data["msg"] = "Note: If there is no such classification information, please search again~, Note: no such classification information, please refer to retrieve ~"
		c.TplName = "jump/error.html"
		return
	}

	o := orm.NewOrm()
	if category.TwoStage != "-" && category.ThreeStage == "-" {
		category_primary := models.Category{}
		o.QueryTable("category").Filter("primary", category.Primary).
			Filter("two_stage", "-").One(&category_primary, "id")
		category.Primary = strconv.Itoa(category_primary.Id)
	}

	if category.ThreeStage != "-" {
		category_primary := models.Category{}
		o.QueryTable("category").Filter("primary", category.Primary).
			Filter("two_stage", "-").One(&category_primary, "id")
		category.Primary = strconv.Itoa(category_primary.Id)

		category_two_stage := models.Category{}
		o.QueryTable("category").Filter("two_stage", category.TwoStage).
			Filter("three_stage", "-").One(&category_two_stage, "id")
		category.TwoStage = strconv.Itoa(category_two_stage.Id)
	}

	is_hidden := c.GetString("is_hidden")
	if is_hidden == "0" {
		category.Is_hidden = true
	} else {
		category.Is_hidden = false
	}

	_, err := o.Update(&category)
	if err != nil {
		c.Data["url"] = "/category_edit"
		c.Data["msg"] = "Failed to change classification information, Failed to change category~"
		c.TplName = "jump/error.html"
		return
	}
	c.Data["url"] = "/category_list"
	c.Data["msg"] = "Change classification information successfully, Change classifieds success~"
	c.TplName = "jump/success.html"
}
