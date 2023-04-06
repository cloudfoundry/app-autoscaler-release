package org.cloudfoundry.autoscaler.scheduler.rest;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.Data;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.BeanFactoryUtils;
import org.springframework.context.ApplicationContext;
import org.springframework.context.ApplicationListener;
import org.springframework.context.event.ContextRefreshedEvent;
import org.springframework.stereotype.Component;
import org.springframework.web.method.HandlerMethod;
import org.springframework.web.servlet.HandlerMapping;
import org.springframework.web.servlet.mvc.method.RequestMappingInfo;
import org.springframework.web.servlet.mvc.method.annotation.RequestMappingHandlerMapping;

import java.util.*;
import java.util.stream.Collectors;

@Slf4j
@Component
public class EndpointsListener implements ApplicationListener<ContextRefreshedEvent> {
    private List<Mapping> mappings = new ArrayList<>();

    public List<Mapping> getMappings() {
        return mappings;
    }

    @Override
    public void onApplicationEvent(ContextRefreshedEvent event) {
        ApplicationContext applicationContext = event.getApplicationContext();
        Map<String, HandlerMapping> allRequestMappings = BeanFactoryUtils.beansOfTypeIncludingAncestors(applicationContext, HandlerMapping.class, true, false);
        for (HandlerMapping handlerMapping : allRequestMappings.values()) {
            if (handlerMapping instanceof RequestMappingHandlerMapping) {
                RequestMappingHandlerMapping requestMappingHandlerMapping = (RequestMappingHandlerMapping) handlerMapping;
                Map<RequestMappingInfo, HandlerMethod> handlerMethods = requestMappingHandlerMapping.getHandlerMethods();

                for (Map.Entry<RequestMappingInfo, HandlerMethod> entry : handlerMethods.entrySet()) {

                    RequestMappingInfo requestMappingInfo = entry.getKey();
                    HandlerMethod handlerMethod = entry.getValue();
                    Mapping mapping = new Mapping();
                    mapping.setMethods(requestMappingInfo.getMethodsCondition().getMethods()
                            .stream().map(Enum::name).collect(Collectors.toSet()));
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
    }
}

