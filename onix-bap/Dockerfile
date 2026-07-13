FROM fidedocker/onix-adapter:1.6.1

RUN apk add --no-cache gettext

COPY templates/ /app/templates/
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENV ENVSUBST_VARS='${REDIS_ADDR} ${NETWORK_PARTICIPANT} ${BAP_ONIX_KEY_ID} ${BAP_ONIX_SIGNING_PRIVATE_KEY} ${BAP_ONIX_SIGNING_PUBLIC_KEY} ${BAP_ONIX_ENCR_PRIVATE_KEY} ${BAP_ONIX_ENCR_PUBLIC_KEY} ${BAP_URL}'

ENTRYPOINT ["/entrypoint.sh"]
CMD ["sh", "-c", "./server --config=/app/config/local-simple-bap.yaml"]
