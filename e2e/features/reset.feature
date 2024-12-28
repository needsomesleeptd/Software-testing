Feature: Reset password with 2FA

Scenario: User reset password with 2FA
  When User send "POST" request to "/login"
  Then the response on /login code should be 200
  And the response on /login should match json:
      """
      {
        "message": "Totp Qr saved on the disk, you must already have totp"
      }
      """
  And user send "POST" request to "/reset_password"
  Then the response on /reset_password code should be 200
  And the response on /reset_password should match json:
      """
      {
        "message": "Password changed successfully"
      }
      """