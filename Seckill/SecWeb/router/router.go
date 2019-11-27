package router

import (
	"Seckill/SecWeb/controller/activity"
	"Seckill/SecWeb/controller/product"
	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/product/list", &product.ProductController{},"*:ListProduct")
	beego.Router("/", &product.ProductController{}, "*:ListProduct")
	beego.Router("/product/create", &product.ProductController{}, "*:CreateProduct")
	beego.Router("/product/submit", &product.ProductController{}, "*:SubmitProduct")

	beego.Router("/activity/create", &activity.ActivityController{}, "*:CreateActivity")
	beego.Router("/activity/list", &activity.ActivityController{}, "*:ListActivity")
	beego.Router("/activity/submit", &activity.ActivityController{}, "*:SubmitActivity")
}
