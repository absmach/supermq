package com.mainflux.loadtest.simulations

object UrlConstants {
  val ManagerUrl = System.getProperty("manager", "http://localhost:8180")
  val HttpAdapterUrl = System.getProperty("http", "http://localhost:8182")
}