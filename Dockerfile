FROM eclipse-temurin:21-jdk AS builder

WORKDIR /app

# Копируем файлы для сборки проекта
COPY gradle/ gradle/
COPY build.gradle.kts settings.gradle.kts gradlew ./
COPY src/ src/

# Предоставляем права на выполнение gradlew
RUN chmod +x ./gradlew

# Собираем проект
RUN ./gradlew build -x test

# Создаем основной образ
FROM eclipse-temurin:21-jre

WORKDIR /app

# Копируем собранный JAR из образа builder
COPY --from=builder /app/build/libs/*.jar app.jar

# Указываем порт, на котором будет работать приложение
EXPOSE 8080

# Команда запуска приложения
ENTRYPOINT ["java", "-jar", "app.jar"]