* /api/v1/
## /api/v1/configuration/

* /config
  * Methods: GET
    * Returns all named configs
    ```
    {
        [
          {
            "config_name": "default,
            "osquery_config: {<snip>}
          },
          {
            "config_name": "default-mac",
            "osquery_config": {<snip>}
          }
        ]
    ```
* config/{config_name}
  * Methods: GET, POST
    * GET: returns the named configed passed in {config_name} in json format
    * POST: Accepts a json post body containing the config specified in the {config_name} parameter.  Note that the passed json config is absolute.  Any parameters not supplied will result in null values being passed in the config.  Be careful not to blow away a config you have already created accidentaly, by only specifying values you want changed.
  * GET - return json configuration of config identified by {config_name}
  * POST - Create/Updated config identified by {config_name}

* /nodes
  * Methods: GET
    * GET: When a post post request is made to this endpoint, it will accept a json blob
with any of the top-level keys specified below.  If any values are not provided,
the existing values of the client will be used (eg not changed)

* /nodes/{node_key}
  * Methods: GET, POST
    * GET: Returns the node configuration of the client specified by {node_key}
    * POST: Accepts a json representation of a client configuration.  This endpoint also accepts a PARTIAL configuration, allowing you to change the values of an individual configuration key, without needed to specify the full configuration to avoid un-setting values


      Data example:
      ```json
        {
          "config_name": "default",
          "node_invalid": false,
          "pending_registration_approval": false,
          "tags: ["tag1", "tag2"]
        }
        ```


* /nodes/{node_key}/approve
  * Methods:  POST
    * POST: This is convenience endpoint to allow easy approval of nodes which have checked in, but have not yet been approved.  This is the equivalent of sending a post a request to the `/node/{node_key}` endpoint with the json body of `{"pending_registration_approval": false}`

* /packs
  * Methods: GET
    * GET: returns a list packs
* /packs/search/{search_string}
  * Methods: GET
    * GET: search packs by name.  simple substring search
* /packs/{pack_name}
  * Methods: POST
    * POST: sets the packqueries for a given pack
      * Data:
          ```json
          {"pack_name": "osx-attacks", "queries": ["OSX_Komplex", "Conduit", "Vsearch"]}
          ```


Node configuration example:
```json
{
  "config_name": "default",
  "node_invalid": false,
  "pending_registration_approval": false,
  "tags: ["tag1", "tag2"]
}
```



## /api/v1/configuration/packs/

* /api/v1/configuration/packs
  * GET - Return list of all packs
* /api/v1/configuration/packs/{pack_name}
  * GET - Return specified pack
* /api/v1/configuration/packs/pack_queries
  * GET - Return list of all pack_queries
* /api/v1/configuration/packs/pack_queries/{query_name}
  * GET - Return query specified by name
  * POST - Update or create specified query by name

## /distributed
The distributed endpoints are used by the osquery nodes and are not intended to be called
by an end-user.  Refer to the osquery documentation for their usage.
* /distributed/read
* /distributed/write


