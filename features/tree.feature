Feature: Directory tree
  As a user I can browse the full directory structure on the telescope.

  Background:
    Given a connected telescope

  Scenario: Tree shows directories and files
    Given the telescope has a directory tree:
      | path              | type   | size   |
      | /user             | dir    | 0      |
      | /user/M42-Orion   | dir    | 0      |
      | /user/M42-Orion/image.fits | file | 4194304 |
    When I request the tree for "/"
    Then the tree should contain directory "user"
    And the tree should contain file "image.fits"

  Scenario: Tree at specific path
    Given the telescope has a directory tree:
      | path                       | type | size    |
      | /user/M42-Orion            | dir  | 0       |
      | /user/M42-Orion/image.fits | file | 4194304 |
    When I request the tree for "/user"
    Then the tree should contain directory "M42-Orion"
