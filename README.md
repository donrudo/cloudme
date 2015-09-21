Cloudme
=======

Testing tool to deploy an application microservices (like in fig or docker-compose)

Compatibility with other container runtimes is planned but docker is the default and only implemented for now

Currently only deploys to a single ``Docker_host``, the spawning of the ``Docker_host`` is planned but not yet ready.

BUILD
=====
  I use gb [http://getgb.io/] for the building process, and the current process I use to build the project is:
```
    cd cloudme_src/
    gb build cloudme 
```

USAGE
=====
  If you want to enable the debugging you can set an environment variable ``DEBUG=*`` and it will show all the output.
``
  cloudme -cmd deploy -config tests/example_app/config.json
``  
or
``
  DEBUG=* cloudme -cmd deploy -config tests/example_app/config.json
``
//  Pre warming steps:
//   1. Validate config file (maestre)
//   2. Dockerclient library requires that the app folder to be converted to .tar for the building process.

//  Execution Steps:
//   0.a TODO: Creates the required Instance (cloudy).
//   0.b TODO: - Setup dependencies (if any)
//
//   1. Build.  (maestre)
//   2. Pull specific commit. (maestre)
//   3. Kill container (if any). (maestre)
//   4. Run container. (maestre)
//   5. TODO:  Test healthcheck. (maestre) (only http 200 checks)
//   6. TODO:  If the healthcheck works register container at the given LB. (cloudy)
 
