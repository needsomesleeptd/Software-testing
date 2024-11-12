import io.gatling.core.Predef._
import io.gatling.http.Predef._
import scala.concurrent.duration._

class ServerPerSecondLoadSimulation extends Simulation {

  val httpProtocolEcho = http.baseUrl("http://echo-hello:8080")
  val httpProtocolGin = http.baseUrl("http://gin-hello:8080")

  val scnEcho = scenario("Echo Server Load Test")
    .exec(http("Echo Metrics per src").get("/"))
  val scnGin = scenario("GinHTTP Server Load Test")
    .exec(http("GinHTTP Metrics per sec").get("/"))

  setUp(
    scnEcho.inject(
        constantUsersPerSec(2000).during(30)
    ).protocols(httpProtocolEcho),
    scnGin.inject(
        constantUsersPerSec(2000).during(30)
    ).protocols(httpProtocolGin)
  ).maxDuration(60.seconds)
}