Feature: List files in an observation
  As a user I can list files inside a specific observation folder.

  Background:
    Given a connected telescope

  Scenario: List files in an observation
    Given observation "M42-Orion" has files:
      | name                          | size   | type |
      | /user/M42-Orion/image001.fits | 4194304 | FITS |
      | /user/M42-Orion/image001.tiff | 2097152 | TIFF |
      | /user/M42-Orion/thumb.jpg     | 65536   | JPEG |
    When I list files for "M42-Orion"
    Then I should get 3 files
    And file "image001.fits" should have type "FITS"

  Scenario: List files in empty observation
    Given observation "Empty" has no files
    When I list files for "Empty"
    Then I should get 0 files
