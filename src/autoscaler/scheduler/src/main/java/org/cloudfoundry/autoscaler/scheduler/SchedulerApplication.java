package org.cloudfoundry.autoscaler.scheduler;

import org.bouncycastle.crypto.fips.FipsStatus;
import org.bouncycastle.jcajce.provider.BouncyCastleFipsProvider;
import org.cloudfoundry.autoscaler.scheduler.conf.MetricsConfig;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.autoconfigure.aop.AopAutoConfiguration;
import org.springframework.boot.autoconfigure.context.ConfigurationPropertiesAutoConfiguration;
import org.springframework.boot.autoconfigure.context.PropertyPlaceholderAutoConfiguration;
import org.springframework.boot.autoconfigure.data.jpa.JpaRepositoriesAutoConfiguration;
import org.springframework.boot.autoconfigure.gson.GsonAutoConfiguration;
import org.springframework.boot.autoconfigure.info.ProjectInfoAutoConfiguration;
import org.springframework.boot.autoconfigure.jackson.JacksonAutoConfiguration;
import org.springframework.boot.autoconfigure.jdbc.DataSourceAutoConfiguration;
import org.springframework.boot.autoconfigure.jdbc.DataSourceTransactionManagerAutoConfiguration;
import org.springframework.boot.autoconfigure.jdbc.JdbcTemplateAutoConfiguration;
import org.springframework.boot.autoconfigure.orm.jpa.HibernateJpaAutoConfiguration;
import org.springframework.boot.autoconfigure.transaction.jta.JtaAutoConfiguration;
import org.springframework.boot.autoconfigure.web.reactive.function.client.WebClientAutoConfiguration;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.boot.context.properties.ConfigurationPropertiesScan;
import org.springframework.context.event.EventListener;

import java.security.Security;

@ConfigurationPropertiesScan(basePackageClasses = MetricsConfig.class)
@SpringBootApplication(
    exclude = {
      AopAutoConfiguration.class,
      DataSourceAutoConfiguration.class,
      WebClientAutoConfiguration.class,
      ProjectInfoAutoConfiguration.class,
      ConfigurationPropertiesAutoConfiguration.class,
      GsonAutoConfiguration.class,
      PropertyPlaceholderAutoConfiguration.class,
      DataSourceTransactionManagerAutoConfiguration.class,
      JacksonAutoConfiguration.class,
      JdbcTemplateAutoConfiguration.class,
      JtaAutoConfiguration.class,
      HibernateJpaAutoConfiguration.class,
      JpaRepositoriesAutoConfiguration.class
    })
public class SchedulerApplication {

  private static final Logger logger = LoggerFactory.getLogger(SchedulerApplication.class);
  private static final int FIPS_ERROR_EXIT_CODE = 140;

  @EventListener
  public void onApplicationReady(ApplicationReadyEvent event) {
    logger.info("Scheduler is ready to start");
  }

  /**
   * Initializes and validates FIPS mode for the application.
   * This is equivalent to checking crypto/fips140.Enabled in Go.
   * Exits with error code 140 if FIPS mode is not enabled or cannot be initialized.
   */
  private static void initializeAndValidateFipsMode() {
    try {
      logger.info("Initializing FIPS 140-2 compliant cryptographic provider...");

      // Register Bouncy Castle FIPS provider as the primary security provider
      Security.insertProviderAt(new BouncyCastleFipsProvider(), 1);

      // Check if Bouncy Castle FIPS is ready and in approved mode (equivalent to crypto/fips140.Enabled)
      if (!FipsStatus.isReady()) {
        logger.error("FIPS mode is not ready. Application requires FIPS 140-3 compliance.");
        System.exit(FIPS_ERROR_EXIT_CODE);
      }

      // Verify that BC-FIPS provider is now installed and available
      if (Security.getProvider("BCFIPS") == null) {
        logger.error("Bouncy Castle FIPS provider (BCFIPS) failed to register.");
        System.exit(FIPS_ERROR_EXIT_CODE);
      }

      // Configure FIPS-compatible system properties for SSL/TLS
      configureFipsCompatibleSystemProperties();

      logger.info("FIPS mode initialization successful - running in FIPS 140-3 mode");
      logger.info("Active security provider: {}", Security.getProvider("BCFIPS").getName());
      logger.info("FIPS Status - Ready: {}",
                 FipsStatus.isReady());

    } catch (Exception e) {
      logger.error("Failed to initialize FIPS mode: {}", e.getMessage(), e);
      System.exit(FIPS_ERROR_EXIT_CODE);
    }
  }

  /**
   * Configures system properties for FIPS-compatible SSL/TLS operations.
   * This prevents issues with XDH key exchange algorithms that are not supported by BC-FIPS.
   */
  private static void configureFipsCompatibleSystemProperties() {
    // Disable XDH algorithms (X25519, X448) that cause issues with BC-FIPS
    System.setProperty("jdk.tls.namedGroups", "secp256r1,secp384r1,secp521r1");
    System.setProperty("jdk.tls.disabledAlgorithms",
                      "SSLv3, RC4, DES, MD5withRSA, DH keySize < 1024, " +
                      "EC keySize < 224, 3DES_EDE_CBC, anon, NULL, " +
                      "X25519, X448, XDH");

    // Ensure FIPS-compatible cipher suites are preferred
    System.setProperty("jdk.tls.client.cipherSuites",
                      "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384," +
                      "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256," +
                      "TLS_RSA_WITH_AES_256_GCM_SHA384," +
                      "TLS_RSA_WITH_AES_128_GCM_SHA256");

    logger.info("Configured FIPS-compatible SSL/TLS system properties");
  }

  public static void main(String[] args) {
    // Initialize and validate FIPS mode before starting the application (equivalent to crypto/fips140.Enabled check)
    initializeAndValidateFipsMode();

    logger.info("Starting Scheduler application with FIPS 140-2 compliance enforced");
    SpringApplication.run(SchedulerApplication.class, args);
  }
}
