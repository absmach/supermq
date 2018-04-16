package com.mainflux.loadtests

import scala.concurrent.duration._
import scalaj.http.Http

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import io.gatling.jdbc.Predef._
import play.api.libs.json._
import play.api.libs.functional.syntax._
import PublishSimulation._

class PublishSimulation extends Simulation {

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

  // Register client
  val clientLocation = Http(s"${BaseUrl}/clients")
    .postData(Client)
    .header(HttpHeaderNames.Authorization, token)
    .header(HttpHeaderNames.ContentType, ContentType)
    .asString
    .headers.get("Location").get(0)
    
  val clientId = clientLocation.split("/")(2)
    
  // Get client key
  val clientRes = Http(s"${BaseUrl}/clients/${clientId}")
    .header(HttpHeaderNames.Authorization, token)
    .header(HttpHeaderNames.ContentType, ContentType)
    .asString
    .body
    
  val clientKey = (Json.parse(clientRes) \ "key").as[String]
    
  // Register channel
  val chanLocation = Http(s"${BaseUrl}/channels")
    .postData(Channel)
    .header(HttpHeaderNames.Authorization, token)
    .header(HttpHeaderNames.ContentType, ContentType)
    .asString
    .headers.get("Location").get(0)
    
  val chanId = chanLocation.split("/")(2)
  
  // Connect client to channel
  Http(s"${BaseUrl}/channels/${chanId}/clients/${clientId}")
    .method("PUT")
    .header(HttpHeaderNames.Authorization, token)
    .asString
    
  // Prepare testing scenario
	val httpProtocol = http
		.baseURL("http://localhost:8182")
		.inferHtmlResources()
		.acceptHeader("*/*")
		.contentTypeHeader("application/json; charset=utf-8")
		.userAgentHeader("curl/7.54.0")

	val scn = scenario("PrepRootUser")
		.exec(http("request_0")
			.post(s"/channels/${chanId}/messages")
			.header(HttpHeaderNames.ContentType, "application/senml+json")
			.header(HttpHeaderNames.Authorization, clientKey)
			.body(RawFileBody("PublishSimulation_0000_request.txt"))
			.check(status.is(202))
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

object PublishSimulation {
  val BaseUrl = "http://localhost:8180"
  val ContentType = "application/json; charset=utf-8"
  
  val User = """{"email":"john.doe@email.com", "password":"123"}"""
  val Client = """{"type":"device", "name":"weio"}"""
  val Channel = """{"name":"mychan"}"""
}
