package org.cloudfoundry.autoscaler.scheduler.rest;

import jakarta.servlet.*;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.codec.binary.Base64;
import org.cloudfoundry.autoscaler.scheduler.conf.HealthServerConfiguration;
import org.springframework.stereotype.Component;
import org.springframework.util.CollectionUtils;
import java.io.IOException;
import java.util.*;
import java.util.stream.Collectors;
import java.util.stream.Stream;

//FIXME Move this class to a suitable package
@Slf4j
@Component
public class MultiPortFilter implements Filter {
    private static Map<String, Boolean> protectedEndpointsMap;

    static {
        protectedEndpointsMap = Map.of("/health/prometheus", true, "/health/liveness", true, "/health/protected", true);
    }

    final HealthServerConfiguration healthServerConfiguration;


    public MultiPortFilter(HealthServerConfiguration healthServerConfiguration){
        this.healthServerConfiguration = healthServerConfiguration;

    }

    @Override
    public void doFilter(ServletRequest servletRequest, ServletResponse servletResponse, FilterChain filterChain) throws IOException, ServletException {
        HttpServletRequest httpRequest = (HttpServletRequest) servletRequest;
        HttpServletResponse httpResponse = (HttpServletResponse) servletResponse;

        //main server - if health endpoints are called, return 404/ error
        // Todo: check for port 8080. The actual port is 6202 (defined in bosh job template)
        //Alternative: Looks for unregister health routes on main server 8080 on the fly
        // example: https://stackoverflow.com/questions/5758504/is-it-possible-to-dynamically-set-requestmappings-in-spring-mvc
        if (!isHealthRequest(httpRequest)) {
            httpResponse.sendError(404, "Health endpoints do not exist");
            return;
        }
        List<String> unprotectedEndpointsConfig = healthServerConfiguration.getUnprotectedEndpoints();

        // Case 1:if Unprotected endpoints are empty or not provided, then protect all health endpoints by default
        if (CollectionUtils.isEmpty(unprotectedEndpointsConfig)) { //health endpoints are authorized
            isUserAuthenticatedOrSendError(filterChain, httpRequest, httpResponse);
            // CASE 2:  // if user configured unprotectedEndpoints, means basic auth is not required to access health endpoints
        } else if (!CollectionUtils.isEmpty(unprotectedEndpointsConfig)) {

            // 1 . Validate provided endpoints with the defined endpoints
            // 2 .  Check if the configured endpoints can be accessible without basic auth

            // if unprotectedEndpointsConfig are not valid, then send unauthorized
            //FIXME refactor validate to suitable class e.g HealthConfig class
            Map<String, Boolean> validateMap = checkValidEndpoints(unprotectedEndpointsConfig);
            if (!CollectionUtils.isEmpty(validateMap)) {
                log.warn("Health configuration: invalid unprotectedEndpoints provided: " + validateMap.get("invalidEndpoints"));
                httpResponse.setHeader("WWW-Authenticate", "Basic");
                httpResponse.sendError(HttpServletResponse.SC_UNAUTHORIZED);
                return;
            }
            // 2 .  Check if the configured endpoints can be accessible without basic auth
            // Match the user configured endpoints (in .properties) with the list of pre-defined protected endpoints
            Map unprotectedConfig = getMapFromList(unprotectedEndpointsConfig);
            List<String> allowedEndpointsWithoutBasicAuth = areEndpointsAuthorized(unprotectedConfig, httpRequest.getRequestURI());
            log.info("Endpoints allowed without basic auth: "+allowedEndpointsWithoutBasicAuth);

            if (!allowedEndpointsWithoutBasicAuth.contains(httpRequest.getRequestURI())) {
                log.warn("Health configuration: Basic auth is required to access protectedEndpoints: "
                        + httpRequest.getRequestURI()+" \nValid unprotected endpoints are: " + allowedEndpointsWithoutBasicAuth);
                httpResponse.setHeader("WWW-Authenticate", "Basic");
                httpResponse.sendError(HttpServletResponse.SC_UNAUTHORIZED);
                return;
            }
            // all set, continue the chain and give access to the resource
            filterChain.doFilter(servletRequest, servletResponse);
        }

    }

    private static Map<String, Boolean> getMapFromList(List<String> unprotectedEndpointsConfig) {
        return unprotectedEndpointsConfig.stream().collect(Collectors.toMap(endpoint -> endpoint, endpoint -> true, (a, b) -> b));
    }


    private static boolean isHealthRequest(HttpServletRequest httpRequest) {
        return httpRequest.getLocalPort() == 8081 && httpRequest.getRequestURI().contains("health");
    }

    //FIXME Move to better class e.g, Authenticator
    private void isUserAuthenticatedOrSendError(FilterChain filterChain, HttpServletRequest httpRequest, HttpServletResponse httpResponse) throws IOException, ServletException {
        final String authorizationHeader = httpRequest.getHeader("Authorization");

        if (healthServerConfiguration.getUsername() == null || healthServerConfiguration.getPassword() == null) {
            log.error("Health configuration: username || password not set");
            httpResponse.setHeader("WWW-Authenticate", "Basic");
            httpResponse.sendError(HttpServletResponse.SC_UNAUTHORIZED);
            return;
        }
        if (authorizationHeader == null) {
            log.error("Basic authentication not provided with the request");
            httpResponse.setHeader("WWW-Authenticate", "Basic");
            httpResponse.sendError(HttpServletResponse.SC_UNAUTHORIZED);
            return;
        }

        String base64Credentials = authorizationHeader.substring("Basic".length()).trim();
        byte[] credDecoded = Base64.decodeBase64(base64Credentials);
        String credentials = new String(credDecoded);
        String[] tokens = credentials.split(":");
        if (tokens.length != 2) {
            log.error("Malformed authorization header");
            httpResponse.sendError(HttpServletResponse.SC_BAD_REQUEST);
            return;
        }
        String username = tokens[0];
        String password = tokens[1];
        
        if (!areBasicAuthCredentialsCorrect(username, password)) {
            httpResponse.setHeader("WWW-Authenticate", "Basic");
            httpResponse.sendError(HttpServletResponse.SC_UNAUTHORIZED);
            return;
        }
        
        if (authorizationHeader != null && isUserAuthenticated(authorizationHeader)) {
            // allow access to health endpoints
            filterChain.doFilter(httpRequest, httpResponse);
        } else {
            httpResponse.setHeader("WWW-Authenticate", "Basic");
            httpResponse.sendError(HttpServletResponse.SC_UNAUTHORIZED);
            return;
        }
    }

    private Map<String, Boolean> checkValidEndpoints(List<String> unprotectedEndpointsConfig) {

        Map<String, Boolean> invalidEndpointsMap = new HashMap<>();
        for (String unprotectedEndpoint : unprotectedEndpointsConfig) {
            if (!protectedEndpointsMap.containsKey(unprotectedEndpoint)) {
                invalidEndpointsMap.put(unprotectedEndpoint, true);
            }
        }
        return invalidEndpointsMap;
    }

    private List<String> areEndpointsAuthorized(Map unprotectedEndpointsConfig, String requestURI) {

        //now check if basic auth is required. then "false" endpoints represents, that resources are mark as unprotected
        Map<String, Boolean> resultUnprotectedEndpoints = new HashMap<>();
        for (Map.Entry<String, Boolean> protectedEndpoint : protectedEndpointsMap.entrySet()) {
            if (unprotectedEndpointsConfig.containsKey(protectedEndpoint.getKey())) {
                resultUnprotectedEndpoints.put(protectedEndpoint.getKey(), false); // rename to unProtectedEndpoints
            }
        }
        // TODO start from here
        List<String> allowedEndpointsWithoutBasicAuth = keys(resultUnprotectedEndpoints, false).toList();
        if (isBasicAuthRequired(requestURI, allowedEndpointsWithoutBasicAuth)) {
            log.info("Endpoints allowed without basic auth: + " + allowedEndpointsWithoutBasicAuth);
            return allowedEndpointsWithoutBasicAuth;
        }
        return allowedEndpointsWithoutBasicAuth;
     }

    private static boolean isBasicAuthRequired(String requestURI, List<String> allowedEndpoints) {
        return !CollectionUtils.isEmpty(allowedEndpoints) && allowedEndpoints.contains(requestURI);
    }

    private <K, V> Stream<K> keys(Map<K, V> map, V val) {
        return map.entrySet().stream().filter(entry -> val.equals(entry.getValue())).map(Map.Entry::getKey);
    }

    private boolean isUserAuthenticated(String authorization) {
        if (healthServerConfiguration.getUsername() == null || healthServerConfiguration.getPassword() == null) {
            log.error("Health configuration: username || password not set");
            return false;
        }
        if (authorization == null) {
            log.error("Basic authentication not provided with the request");
            return false;
        }

        String base64Credentials = authorization.substring("Basic".length()).trim();
        byte[] credDecoded = Base64.decodeBase64(base64Credentials);
        String credentials = new String(credDecoded);
        String[] tokens = credentials.split(":");
        if (tokens.length != 2) {
            log.error("Malformed authorization header");
            return false;
        }
        String username = tokens[0];
        String password = tokens[1];
        return areBasicAuthCredentialsCorrect(username, password);
    }

    private boolean areBasicAuthCredentialsCorrect(String username, String password) {
        return healthServerConfiguration.getUsername().equals(username) && healthServerConfiguration.getPassword().equals(password);
    }

   /* private List<Mapping> getAllRequestMappings() {
        List<Mapping> mappings = new ArrayList<>();
        Map<String, HandlerMapping> allRequestMappings = BeanFactoryUtils.beansOfTypeIncludingAncestors(webApplicationContext, HandlerMapping.class, true, false);
        for (HandlerMapping handlerMapping : allRequestMappings.values()) {
            if (handlerMapping instanceof RequestMappingHandlerMapping) {
                RequestMappingHandlerMapping requestMappingHandlerMapping = (RequestMappingHandlerMapping) handlerMapping;
                Map<RequestMappingInfo, HandlerMethod> handlerMethods = requestMappingHandlerMapping.getHandlerMethods();

                for (Map.Entry<RequestMappingInfo, HandlerMethod> entry : handlerMethods.entrySet()) {

                    RequestMappingInfo requestMappingInfo = entry.getKey();
                    HandlerMethod handlerMethod = entry.getValue();
                    Mapping mapping = new Mapping();
                    mapping.setMethods(requestMappingInfo.getMethodsCondition().getMethods().stream().map(Enum::name).collect(Collectors.toSet()));
                    mapping.setPatterns(requestMappingInfo.getPatternsCondition().getPatterns());
                    Arrays.stream(handlerMethod.getMethodParameters()).forEach(methodParameter -> {
                        mapping.getParams().add(methodParameter.getParameter().getType().getSimpleName());
                    });
                    mappings.add(mapping);
                }
                mappings.sort(Comparator.comparing(o -> o.getPatterns().stream().findFirst().orElse("")));
                try {
                    log.info(new ObjectMapper().writeValueAsString(mappings));
                } catch (JsonProcessingException e) {
                    throw new RuntimeException(e);
                }
            }
        }
        return mappings;
    }


    @Data
    public class Mapping {
        private Set<String> patterns;
        private Set<String> methods;
        private List<String> params;

        public List<String> getParams() {
            if (params == null) {
                params = new ArrayList<>();
            }
            return params;

        }
    }*/

}


