package org.cloudfoundry.autoscaler.scheduler.rest.health;


import io.prometheus.client.CollectorRegistry;
import io.prometheus.client.exporter.common.TextFormat;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.codec.binary.Base64;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.OutputStreamWriter;
import java.util.HashMap;
import java.util.Map;

@Slf4j
@RestController
@RequestMapping(value = {"/health"})
public class HealthRestController {
    @GetMapping(value = "/prometheus")
    public ResponseEntity<String> getPrometheusMetrics()
            throws IOException {

        final ByteArrayOutputStream stream = new ByteArrayOutputStream();
        try (OutputStreamWriter writer = new OutputStreamWriter(stream)) {
            TextFormat.write004(writer, CollectorRegistry.defaultRegistry.metricFamilySamples());
        }
        return new ResponseEntity<>(stream.toString(), HttpStatus.OK);
    }

    @GetMapping(value = "/liveness")
    public ResponseEntity<Map<String, Object>> getLiveness() {
        Map<String, Object> body = new HashMap<>();
        body.put("status", "Up");
        return new ResponseEntity<>(body, HttpStatus.OK);
    }



}




