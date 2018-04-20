import sbt.io.Path._
import sbt.io.PathFinder._

enablePlugins(JavaAppPackaging)

name := "load-test"
version := "1.0-SNAPSHOT"

scalaVersion := "2.12.4"

val gatlingVersion = "2.3.1"
val circeVersion = "0.9.3"

libraryDependencies ++= Seq(
  "io.gatling.highcharts" %  "gatling-charts-highcharts" % gatlingVersion,
  "io.gatling"            %  "gatling-test-framework"    % gatlingVersion,
  "org.scalaj"            %% "scalaj-http"               % "2.3.0",
  "io.circe"              %% "circe-core"                % circeVersion,
  "io.circe"              %% "circe-generic"             % circeVersion,
  "io.circe"              %% "circe-parser"              % circeVersion
)

mainClass in Compile := Some("com.mainflux.loadtest.Engine")

mappings in Universal ++= {
  val binariesDir = target.value / "scala-2.12" / "classes"
  (binariesDir.allPaths --- binariesDir) pair relativeTo(baseDirectory.value)
}
