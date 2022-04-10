FROM fluent/fluentd:edge

USER root

RUN apk add --no-cache --update --virtual .build-deps \
        sudo build-base ruby-dev \
 && sudo gem install 'elasticsearch:7.17.0' \
                     'elasticsearch-api:7.17.0' \
                     'elasticsearch-xpack:7.17.0' \
                     fluent-plugin-elasticsearch \
                     oj \
                     fluent-plugin-rewrite-tag-filter \
 && sudo gem sources --clear-all \
 && apk del .build-deps \
 && rm -rf /tmp/* /var/tmp/* /usr/lib/ruby/gems/*/cache/*.gem

USER fluent