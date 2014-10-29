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
    RUN  curl -L https://github.com/docker-infra/reefer/releases/download/v0.0.1/reefer.gz | \
         gunzip > /usr/bin/reefer && chmod a+x /usr/bin/reefer
    ADD  templates /
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
