package org.cloudfoundry.autoscaler.scheduler.rest;

import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.extern.slf4j.Slf4j;
import org.springframework.web.context.request.WebRequestInterceptor;
import org.springframework.web.method.HandlerMethod;
import org.springframework.web.servlet.handler.WebRequestHandlerInterceptorAdapter;
@Slf4j
public class HandlerInterceptor extends WebRequestHandlerInterceptorAdapter {
    public HandlerInterceptor(WebRequestInterceptor requestInterceptor) {
        super(requestInterceptor);
    }
    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) {

        String handlerPackageName = ((HandlerMethod) handler)
                .getBean().getClass().getPackage().getName();

        if(request.getLocalPort() == 8081) {
           log.info("server on 8081");
            return true;
        }

        if(request.getLocalPort() == 8080 ) {
            log.info("server on 8080");
            return true;
        }

        response.setStatus(404);
        return false;
    }

}
