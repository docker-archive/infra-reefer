# Reefer
*Managing a stable environment in your container.*

![refrigerated container](http://upload.wikimedia.org/wikipedia/commons/6/60/Redundantreefer.JPG)

Reefer is used to render templates based on environment variables
before exec'ing a given process.

This is useful to configure legacy applications by environment
variables to support [12factor app like configs](http://12factor.net/config).

# Example: nginx + ssl certificates on Docker
First we create a image with nginx and reefer using a Dockerfile like this:

    FROM nginx
    RUN  curl -L https://github.com/docker-infra/reefer/releases/download/v0.0.3/reefer.gz | \
         gunzip > /usr/bin/reefer && chmod a+x /usr/bin/reefer
    COPY  templates /
    ENTRYPOINT [ "/usr/bin/reefer", \
      "-t", "/templates/nginx.conf.tmpl:/etc/nginx/nginx.conf", \
      "-t", "/templates/cert.pem.tmpl:/etc/nginx/cert.pem", \
      "-t",  "/templates/key.pem.tmpl:/etc/nginx/key.pem", \
      "-t",  "/templates/htpasswd.tmpl:/etc/nginx/htpasswd", \
      "/usr/bin/nginx", "-g", "daemon off;"
    ]

The files in the templates/ directory would look something like this:

cert.pem.tmpl:

    {{ .Env "TLS_CERT" }}


key.pem.tmpl:

    {{ .Env "TLS_KEY" }}


nginx.conf.tmpl:

    http {
      server {
        listen 443;
      
        ssl    on;
        ssl_certificate     /etc/nginx/cert.pem;
        ssl_certificate_key /etc/nginx/key.pem;
      
        server_name {{ .Env "DOMAIN" }};
        location / {
          auth_basic "secret";
          auth_basic_user_file /etc/nginx/htpasswd;
  
          root   /srv/www/htdocs;
          index  index.html;
        }
      }
    }


htpasswd.tmpl:

    alice:{{ .Env "PASS_ALICE" }}
    bob:{{ .Env "PASS_BOB" }}


Now you can start the image like this:

    $ docker run -e TLS_CERT=`cat your-cert.pem` -e TLS_KEY=`cat your-key.pem` \
      -e DOMAIN=example.com -e PASS_ALICE=foobar23 -e PASS_BOB=blafasel -p 443:443 your-image

Reefer will read the environment variables and render the templates.
After that, it will exec() the remaining parameter (nginx -g daemon off; in this example).

# Passing environment variable through to your application

By default, reefer will not "pass through" environment variables you set to the application it executes; they will be available for the templates, but not the application. Notable exceptions:

    COLORS
    DISPLAY
    HOME
    HOSTNAME
    KRB5CCNAME
    LS_COLORS
    PATH
    PS1
    PS2
    TZ
    XAUTHORITY
    XAUTHORIZATION

This is done for security reasons because one of the primary uses of reefer is to pass sensitive information (private keys, etc.) used in the generation of your templates. It is generally a good idea to not have these environment variables "floating around" in the container environment. If you would like to pass through environment variables to other applications in your container, you can specify individual environment variables to "keep" with `-e` like so:

    ENTRYPOINT [ "/usr/bin/reefer", \
      "-t", "/templates/app.conf.tmpl:/app/etc/app.conf", \
      "-t", "/templates/cert.pem.tmpl:/app/certs/cert.pem", \
      "-t",  "/templates/key.pem.tmpl:/app/certs/key.pem", \
      "-e",  "IMPORTANT_CONFIG_VAR", \
      "/app/app"
    ]
    
You can pass ALL environment variables through with `-E`:

    ENTRYPOINT [ "/usr/bin/reefer", \
      "-t", "/templates/app.conf.tmpl:/app/etc/app.conf", \
      "-E", \
      "/app/app"
    ]

