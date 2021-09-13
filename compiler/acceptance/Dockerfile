# TFX build 
FROM tensorflow/tfx:1.2.0 as tfx_base
# Main 
FROM tensorflow/tensorflow:2.4.0-gpu 
COPY --from=tfx_base tfx/src/ tfx-src 

RUN pip install tfx==1.2.0
RUN pip install kfp
COPY module.py src/
COPY pipeline.py src/

WORKDIR /src 

ENV PYTHONPATH="/src:${PYTHONPATH}"
ENV PYTHONPATH="/tfx-src:${PYTHONPATH}"


ENTRYPOINT ["python3", "/tfx-src/tfx/scripts/run_executor.py"]