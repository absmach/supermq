package com.mainflux.loadtests

import scala.concurrent.duration._
import scalaj.http.Http

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import io.gatling.jdbc.Predef._
import play.api.libs.json._
import play.api.libs.functional.syntax._
import CreateAndRetrieveClientSimulation._

class CreateAndRetrieveClientSimulation extends Simulation {

  // Register user
  Http(s"${BaseUrl}/users")
    .postData(User)
    .header(HttpHeaderNames.ContentType, ContentType)
    .asString
  
  // Login user
  val tokenRes = Http(s"${BaseUrl}/tokens")
    .postData(User)
    .header(HttpHeaderNames.ContentType, ContentType)
    .asString
    .body
    
  val token = (Json.parse(tokenRes) \ "token").as[String]

  // Prepare testing scenario
	val httpProtocol = http
		.baseURL(BaseUrl)
		.inferHtmlResources()
		.acceptHeader("*/*")
		.contentTypeHeader(ContentType)
		.userAgentHeader("curl/7.54.0")

	val scn = scenario("PrepRootUser")
		.exec(http("request_0")
			.post("/clients")
			.header(HttpHeaderNames.ContentType, ContentType)
			.header(HttpHeaderNames.Authorization, token)
			.body(RawFileBody("CreateAndRetrieveClientSimulation_0000_request.txt"))
			.check(status.is(201))
			.check(headerRegex(HttpHeaderNames.Location, "(.*)").saveAs("location"))
		)
		.exec(http("request_1")
		  .get("${location}")
		  .header(HttpHeaderNames.Authorization, token)
		  .check(status.is(200))
		)

	setUp(
	  scn.inject(
	    constantUsersPerSec(100) during (15 second),
	    constantUsersPerSec(250) during (15 second),
	    constantUsersPerSec(500) during (15 second),
	    constantUsersPerSec(750) during (15 second),
	    constantUsersPerSec(1000) during (15 second)
	  )
	).protocols(httpProtocol)
}

object CreateAndRetrieveClientSimulation {
  val BaseUrl = "http://localhost:8180"
  val ContentType = "application/json; charset=utf-8"
  
  val User = """{"email":"john.doe@email.com", "password":"123"}"""
}