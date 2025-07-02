# PromQL 익스포터

5G/CNF(Cloud Native Functions) 모니터링을 위한 Go 기반 Prometheus 메트릭 익스포터입니다. CSV 파일, Kubernetes 클러스터, HTTP 엔드포인트 등 다양한 소스에서 메트릭을 수집하여 Prometheus 형식으로 노출합니다.

## 개요

이 프로젝트는 5G 코어 네트워크 기능(CNF) 모니터링을 위해 특별히 설계된 Prometheus 익스포터로, 특히 Samsung CPC(제어 평면 구성 요소) 메트릭에 중점을 둡니다. 다양한 소스에서 메트릭을 수집하고 내보냅니다:

- OSS(Operation Support System)의 CSV 기반 메트릭
- Kubernetes 클러스터 API를 통한 메트릭
- HTTP 엔드포인트 스크래핑
- 실시간 성능 데이터 집계

## 아키텍처

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   OSS 시스템     │    │  Kubernetes API  │    │  HTTP 엔드포인트 │
│   (CSV 파일)    │    │                  │    │                 │
└─────────┬───────┘    └────────┬─────────┘    └─────────┬───────┘
          │                     │                        │
          └─────────────────────┼────────────────────────┘
                                │
                    ┌───────────▼──────────┐
                    │  PromQL 익스포터      │
                    │  - 데이터 수집       │
                    │  - 메트릭 처리       │
                    │  - 파일 관리         │
                    └───────────┬──────────┘
                                │
                    ┌───────────▼──────────┐
                    │  Prometheus 형식     │
                    │  HTTP 엔드포인트     │
                    │  /metrics, /api/...  │
                    └──────────────────────┘
```

## 주요 기능

- **다중 소스 데이터 수집**: CSV 파일, Kubernetes API, HTTP 엔드포인트 지원
- **자동 파일 관리**: 수집된 데이터의 자동 백업 및 정리
- **5G/CNF 메트릭**: 5G 코어 네트워크 기능을 위한 전문 메트릭
- **Kubernetes 통합**: OpenShift/Kubernetes 클러스터 모니터링 기본 지원
- **RESTful API**: 메트릭 접근을 위한 API 엔드포인트 제공
- **컨테이너화 배포**: 설정 가능한 매개변수로 Docker 지원

## 프로젝트 구조

```
PromQL-exporter/
├── cmd/
│   └── exporter.go          # 메인 애플리케이션 진입점
├── pkg/
│   ├── exporter/            # 핵심 익스포터 로직
│   │   ├── appCollector.go  # 애플리케이션 메트릭 수집기
│   │   └── model.go         # 데이터 모델 및 구조체
│   ├── csv/                 # CSV 파일 처리
│   ├── curl/                # HTTP 클라이언트 유틸리티
│   ├── k8sClient/           # Kubernetes 클라이언트
│   ├── metricApi/           # API 핸들러
│   └── utils/               # 유틸리티 함수
├── cfg/                     # 구성 관리
├── logger/                  # 로깅 유틸리티
├── config.yml               # 메인 설정
├── app_config.yml           # 애플리케이션별 메트릭 설정
├── cnf_config.yml           # CNF 메트릭 설정
├── Dockerfile               # 컨테이너 빌드 설정
└── Makefile                 # 빌드 자동화
```

## 설정

### 메인 설정 (`config.yml`)

```yaml
logging:
  LEVEL: INFO
  ENCODE: json

file:
  MEC_CONFIG: "/mnt/data/config"
  CSV_PATH: "/mnt/data/exporter"
  API_PATH: "/mnt/data/exporter/api"
  FAMILY_NAME: ["UECON_AMF", "TMSI_AMF", ...]

exporter:
  CURL_URL: "https://your-oss-system/oss/performanceData"
  OSS_USERNAME: "username"
  OSS_PASSWORD: "password"
```

### 애플리케이션 메트릭 (`app_config.yml`)

다양한 엔드포인트에서 메트릭 수집을 정의:

```yaml
cpu/metrics:
  collects:
  - metrics:
      node_cpu_seconds_total:
        type: counter
        prefix: p5g_wrcp1
        description: wrcp1_core_cpu_value
        url: "http://prometheus-api/api/v1/query?query=node_cpu_seconds_total"
```

### CNF 메트릭 (`cnf_config.yml`)

5G CNF 전용 메트릭 설정:

```yaml
metrics:
  amf_ue_connect_attempt_count:
    type: counter
    description: "UECON_AMF"
    labels: ["ne_id", "system_id", "ne_name", ...]
    value_sequence: 7
```

## 설치 및 배포

### Docker 빌드

```bash
# 이미지 빌드
make save

# 또는 수동 빌드
docker build -t promql-exporter:latest .
```

### 로컬 개발

```bash
# 의존성 설치
go mod download

# 애플리케이션 빌드
go build -o bin/exporter ./cmd/exporter.go

# 설정과 함께 실행
./bin/exporter -metricConfig cnf_config.yml -config-metrics app_config.yml
```

### Kubernetes 배포

```bash
# Helm을 사용한 배포
helm install promql-exporter ./chart --values values.yaml
```

## 사용법

### 익스포터 실행

```bash
# 기본 설정
./exporter

# 사용자 정의 설정 파일
./exporter -metricConfig custom_cnf.yml -config-metrics custom_app.yml
```

### API 엔드포인트

- `GET /metrics` - Prometheus 메트릭 엔드포인트
- `GET /api/metrics` - CNF 메트릭 API
- `GET /cpu/metrics` - CPU 메트릭 엔드포인트
- `GET /mem/metrics` - 메모리 메트릭 엔드포인트
- `GET /pod/metrics` - Pod 메트릭 엔드포인트

### 샘플 메트릭 출력

```prometheus
# HELP p5g_exporter_amf_ue_connect_attempt_count AMF UE 연결 시도
# TYPE p5g_exporter_amf_ue_connect_attempt_count counter
p5g_exporter_amf_ue_connect_attempt_count{ne_id="amf-001",location="datacenter-1"} 1500

# HELP p5g_mec_node_cpu_seconds_total MEC CPU 초 총합
# TYPE p5g_mec_node_cpu_seconds_total counter
p5g_mec_node_cpu_seconds_total{container="prometheus",cpu="0",instance="node-1"} 12345.67
```

## 모니터링 메트릭 범주

### 5G 코어 네트워크 기능
- **AMF (Access and Mobility Management Function - 접근 및 이동성 관리 기능)**
  - UE 연결 시도/성공
  - TMSI (임시 모바일 가입자 식별자) 작업
  - 트랜잭션 처리 통계

- **SMF (Session Management Function - 세션 관리 기능)**
  - GTP-C 터널 엔드포인트 작업
  - 세션 설정 메트릭
  - 트랜잭션 처리량

- **UPF (User Plane Function - 사용자 평면 기능)**
  - 패킷 검사 통계
  - 데이터 포워딩 트래픽
  - PDCP 볼륨 및 패킷 메트릭

### 인프라 메트릭
- **CPU 사용률**: 클러스터 전반의 노드 레벨 CPU 메트릭
- **메모리 사용량**: 메모리 소비 및 가용성
- **Pod 메트릭**: 컨테이너 리소스 사용률
- **네트워크 메트릭**: 에어 인터페이스 MAC 패킷 통계

## 데이터 흐름

1. **수집 단계**:
   - HTTP를 통해 OSS에서 성능 데이터 가져오기
   - 인프라 메트릭을 위한 Kubernetes API 쿼리
   - 과거 데이터를 위한 CSV 파일 처리

2. **처리 단계**:
   - 메트릭 데이터 파싱 및 검증
   - 메트릭 타입 변환 적용
   - Prometheus 호환 라벨 생성

3. **내보내기 단계**:
   - HTTP 엔드포인트를 통한 메트릭 노출
   - 처리된 파일 백업
   - 데이터 보존 정책 유지

## 개발

### 전제 조건

- Go 1.20+
- Docker
- Kubernetes 클러스터 접근 (선택사항)
- CSV 데이터를 위한 OSS 시스템 접근

### 빌드

```bash
# 다양한 플랫폼을 위한 크로스 컴파일
make cc        # Windows
make cclinux   # Linux ARM/ARM64

# 버전 관리가 포함된 Docker 이미지
make save tag=v1.0.0
```

### 새 메트릭 추가

1. `cnf_config.yml`에서 메트릭 정의:
```yaml
new_metric_name:
  type: gauge
  description: "FAMILY_NAME"
  labels: ["label1", "label2"]
  value_sequence: 10
```

2. `config.yml`에 패밀리 이름 추가:
```yaml
FAMILY_NAME: [..., "NEW_FAMILY_NAME"]
```

3. 적절한 수집기에서 수집 로직 구현

## 문제 해결

### 일반적인 문제

1. **CSV 파일을 찾을 수 없음**
   - 설정에서 파일 경로 확인
   - OSS 시스템 연결 확인
   - 적절한 파일 권한 확인

2. **Kubernetes 인증**
   - 서비스 계정 권한 확인
   - 클러스터 연결 확인
   - 토큰 생성 검증

3. **메모리 문제**
   - 대용량 CSV 파일 처리 모니터링
   - 백업 간격 조정
   - 데이터 경로의 디스크 공간 확인

### 로그

애플리케이션은 구조화된 JSON 로깅을 사용:

```bash
# 컨테이너에서 로그 확인
docker logs promql-exporter

# 로그 레벨별 필터링
docker logs promql-exporter 2>&1 | grep ERROR
```

## 보안 고려사항

- 민감한 URL과 자격 증명은 설정 예제에서 마스킹됨
- Kubernetes API 접근을 위한 베어러 토큰 인증 사용
- HTTP 클라이언트에 대한 TLS 검증 설정 가능
- 데이터 디렉토리에 적절한 파일 권한 설정

## 성능

- 설정 가능한 간격으로 메트릭 처리
- 대용량 데이터셋을 위한 파일 기반 캐싱 구현
- 동시 메트릭 수집 지원
- 컨테이너화된 배포에 최적화

## 라이선스

이 소프트웨어는 KT Corp의 독점 소프트웨어입니다. 모든 권리 보유. 라이선스 계약에 따라 사용이 제한됩니다.

---

**참고**: 이 익스포터는 Samsung CPC 5G 네트워크 모니터링을 위해 특별히 설계되었으며, 다른 환경에서는 커스터마이징이 필요할 수 있습니다.