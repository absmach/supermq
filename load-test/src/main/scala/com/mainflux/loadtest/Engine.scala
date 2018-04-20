package com.mainflux.loadtest

import io.gatling.app.Gatling
import io.gatling.core.config.GatlingPropertiesBuilder
import scala.util.Properties.envOrElse

object Engine extends App {
  val ManagerUrl = envOrElse("MF_LT_MANAGER_URL", "http://localhost:8180")
  val HttpAdapterUrl = envOrElse("MF_LT_HTTP_ADAPTER_URL", "http://localhost:8182")
  
  val props = new GatlingPropertiesBuilder()
    .resultsDirectory("target/gatling")
    .binariesDirectory("target/scala-2.12/classes")

  Gatling.fromMap(props.build)
}