package bloomz

// This is a wrapper to statisfy the GRPC service interface
// It is meant to be used by the main executable that is the server for the specific backend type (falcon, gpt3, etc)
import (
	"fmt"

	"github.com/go-skynet/LocalAI/pkg/grpc/base"
	pb "github.com/go-skynet/LocalAI/pkg/grpc/proto"
	"github.com/rs/zerolog/log"

	"github.com/go-skynet/bloomz.cpp"
)

type LLM struct {
	base.Base

	bloomz *bloomz.Bloomz
}

func (llm *LLM) Load(opts *pb.ModelOptions) error {
	if llm.Base.State != pb.StatusResponse_UNINITIALIZED {
		log.Warn().Msgf("bloomz backend loading %s while already in state %s!", opts.Model, llm.Base.State.String())
	}

	llm.Base.Lock()
	defer llm.Base.Unlock()
	model, err := bloomz.New(opts.ModelFile)
	llm.bloomz = model
	return err
}

func buildPredictOptions(opts *pb.PredictOptions) []bloomz.PredictOption {
	predictOptions := []bloomz.PredictOption{
		bloomz.SetTemperature(float64(opts.Temperature)),
		bloomz.SetTopP(float64(opts.TopP)),
		bloomz.SetTopK(int(opts.TopK)),
		bloomz.SetTokens(int(opts.Tokens)),
		bloomz.SetThreads(int(opts.Threads)),
	}

	if opts.Seed != 0 {
		predictOptions = append(predictOptions, bloomz.SetSeed(int(opts.Seed)))
	}

	return predictOptions
}

func (llm *LLM) Predict(opts *pb.PredictOptions) (string, error) {
	llm.Base.Lock()
	defer llm.Base.Unlock()

	return llm.bloomz.Predict(opts.Prompt, buildPredictOptions(opts)...)
}

// fallback to Predict
func (llm *LLM) PredictStream(opts *pb.PredictOptions, results chan string) error {
	llm.Base.Lock()

	go func() {
		res, err := llm.bloomz.Predict(opts.Prompt, buildPredictOptions(opts)...)

		if err != nil {
			fmt.Println("err: ", err)
		}
		results <- res
		close(results)
		llm.Base.Unlock()
	}()

	return nil
}
