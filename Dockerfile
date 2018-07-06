FROM scratch
COPY steam-authority /steam-authority
COPY templates /templates
COPY node_modules /node_modules
COPY assets /assets
EXPOSE 8085
CMD ["/steam-authority"]
