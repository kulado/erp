package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/kulado/erp/models"
	"github.com/kulado/erp/plugins/permission"
	"html/template"
)

type BrandController struct {
	beego.Controller
}

// Trademark list page
func (c *BrandController) Get() {
	brand := []models.Brand{}
	o := orm.NewOrm()
	o.QueryTable("brand").All(&brand)

	c.Data["brand"] = brand
	c.Layout = "common.tpl"
	c.TplName = "brand/brand_list.html"
}

// Add a trademark page, Adding trademarks page
func (c *BrandController) Brand_add() {
	if !permission.GetOneItemPermission(c.GetSession("username").(string), "AddBrand") {
		c.Abort("401")
	}
	c.Data["xsrfdata"] = template.HTML(c.XSRFFormHTML())
	c.Layout = "common.tpl"
	c.TplName = "brand/brand_add.html"
}

// Add trademark post submission, Adding to submit trademark post
func (c *BrandController) Brand_add_post() {
	if !permission.GetOneItemPermission(c.GetSession("username").(string), "AddBrand") {
		c.Abort("401")
	}
	brand := models.Brand{}
	brand.Name = c.GetString("name")

	o := orm.NewOrm()
	exit := o.QueryTable("brand").Filter("name", brand.Name).Exist()
	if exit {
		c.Data["msg"] = "This brand name already exists. Please do not add it repeatedly~, This brand name already exists, do not repeat add ~"
		c.Data["url"] = "/brand_add"
		c.TplName = "jump/error.html"
		return
	} else {
		_, err := o.Insert(&brand)
		if err != nil {
			logs.Error(c.GetSession("uid"), "Add brand error: ", err)
			c.Data["msg"] = "Add failed, please try again later or contact administrator~, Add failed, please try again later or contact your administrator ~"
			c.Data["url"] = "/brand_add"
			c.TplName = "jump/error.html"
			return
		} else {
			c.Data["msg"] = "Add brand " + brand.Name + " success~"
			c.Data["url"] = "/brand_list"
			c.TplName = "jump/success.html"
			return
		}
	}
}
