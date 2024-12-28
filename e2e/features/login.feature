Feature: Login with 2FA

Scenario: User login with 2FA
  When User send "POST" request to "/login"
  Then the response on /login code should be 200
  And the response on /login should match json:
      """
      {
        "message": "Totp Qr saved on the disk, you must already have totp"
      }
      """
  And user send "POST" request to "/verify"
  Then the response on /verify code should be 200
  And the response on /verify should match json:
      """
      {
        "message": "Auth success!"
      }
      """